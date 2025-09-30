package repository

import (
	"context"
	"database/sql" // Changed from go.mongodb.org/mongo-driver/mongo
	"engagementService/internal/model"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	"time"

	commonErrors "engagementService/internal/errors" // Import the new errors package
)

type SubscriptionRepo interface {
	EnsureIndexes(ctx context.Context) error

	Create(ctx context.Context, s *model.Subscription) error
	Delete(ctx context.Context, followerID, followeeID uuid.UUID) error
	HardDelete(ctx context.Context, followerID, followeeID uuid.UUID) error

	IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error)

	GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error)

	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
}

// PostgresSubscriptionRepo replaces MongoSubscriptionRepo
type PostgresSubscriptionRepo struct {
	db *sql.DB // Changed from *mongo.Collection
}

// NewPostgresSubscriptionRepo replaces NewMongoSubscriptionRepo
func NewPostgresSubscriptionRepo(db *sql.DB) *PostgresSubscriptionRepo {
	return &PostgresSubscriptionRepo{
		db: db,
	}
}

func (r *PostgresSubscriptionRepo) EnsureIndexes(ctx context.Context) error {
	// Create table if it does not exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY,
		follower_id UUID NOT NULL,
		followee_id UUID NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE (follower_id, followee_id)
	);`
	_, err := r.db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return err
	}

	// Create indexes
	// The unique index is already part of the table creation
	// For followee_id and follower_id, we can create non-unique indexes for performance
	createFolloweeIndexSQL := `CREATE INDEX IF NOT EXISTS i_followee ON subscriptions (followee_id, deleted_at);`
	_, err = r.db.ExecContext(ctx, createFolloweeIndexSQL)
	if err != nil {
		return err
	}

	createFollowerIndexSQL := `CREATE INDEX IF NOT EXISTS i_follower ON subscriptions (follower_id, deleted_at);`
	_, err = r.db.ExecContext(ctx, createFollowerIndexSQL)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresSubscriptionRepo) Create(ctx context.Context, s *model.Subscription) error {
	if s == nil {
		return errors.New("subscription is nil")
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.DeletedAt = nil // Ensure deleted_at is nil for new subscriptions

	insertSQL := `
		INSERT INTO subscriptions (id, follower_id, followee_id, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5);`

	_, err := r.db.ExecContext(ctx, insertSQL, s.ID, s.FollowerID, s.FolloweeID, s.CreatedAt, s.DeletedAt)
	if err != nil {
		// Check for duplicate key error (PostgreSQL-specific error code)
		// This is a common way to check for unique constraint violations in pq driver
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return commonErrors.ErrDuplicateSubscription // Use commonErrors
		}
		return err
	}
	return nil
}

// Soft delete: set deleted_at if not already deleted
func (r *PostgresSubscriptionRepo) Delete(ctx context.Context, followerID, followeeID uuid.UUID) error {
	now := time.Now().UTC()
	updateSQL := `
		UPDATE subscriptions
		SET deleted_at = $1
		WHERE follower_id = $2 AND followee_id = $3 AND deleted_at IS NULL;`

	result, err := r.db.ExecContext(ctx, updateSQL, now, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return commonErrors.ErrNotFound // Use commonErrors
	}
	return nil
}

func (r *PostgresSubscriptionRepo) HardDelete(ctx context.Context, followerID, followeeID uuid.UUID) error {
	deleteSQL := `
		DELETE FROM subscriptions
		WHERE follower_id = $1 AND followee_id = $2;`

	result, err := r.db.ExecContext(ctx, deleteSQL, followerID, followeeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return commonErrors.ErrNotFound // Use commonErrors
	}
	return nil
}

func (r *PostgresSubscriptionRepo) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	querySQL := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE follower_id = $1 AND followee_id = $2 AND deleted_at IS NULL;`

	var count int
	err := r.db.QueryRowContext(ctx, querySQL, followerID, followeeID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresSubscriptionRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error) {
	querySQL := `
		SELECT id, follower_id, followee_id, created_at, deleted_at
		FROM subscriptions
		WHERE followee_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;`

	rows, err := r.db.QueryContext(ctx, querySQL, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Subscription
	for rows.Next() {
		var s model.Subscription
		var deletedAt sql.NullTime // Use sql.NullTime for nullable TIMESTAMP
		err := rows.Scan(&s.ID, &s.FollowerID, &s.FolloweeID, &s.CreatedAt, &deletedAt)
		if err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			s.DeletedAt = &deletedAt.Time
		} else {
			s.DeletedAt = nil
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresSubscriptionRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int64) ([]model.Subscription, error) {
	querySQL := `
		SELECT id, follower_id, followee_id, created_at, deleted_at
		FROM subscriptions
		WHERE follower_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;`

	rows, err := r.db.QueryContext(ctx, querySQL, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Subscription
	for rows.Next() {
		var s model.Subscription
		var deletedAt sql.NullTime
		err := rows.Scan(&s.ID, &s.FollowerID, &s.FolloweeID, &s.CreatedAt, &deletedAt)
		if err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			s.DeletedAt = &deletedAt.Time
		} else {
			s.DeletedAt = nil
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresSubscriptionRepo) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	querySQL := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE followee_id = $1 AND deleted_at IS NULL;`

	var count int64
	err := r.db.QueryRowContext(ctx, querySQL, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresSubscriptionRepo) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	querySQL := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE follower_id = $1 AND deleted_at IS NULL;`

	var count int64
	err := r.db.QueryRowContext(ctx, querySQL, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
