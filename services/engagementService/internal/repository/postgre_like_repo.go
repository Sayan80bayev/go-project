package repository

import (
	"context"
	"database/sql"
	commonErrors "engagementService/internal/errors"
	"engagementService/internal/model"
	"errors"
	"fmt"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	// Removed "go.uber.org/zap"
	"time"
)

// LikeRepo defines the interface for like-related database operations.
type LikeRepo interface {
	Create(ctx context.Context, s *model.Like) (*model.Like, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Like, error)
	GetByUserID(ctx context.Context, id uuid.UUID, limit, offset int) ([]*model.Like, error)
	GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*model.Like, error)
	Delete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error
}

// PostgresLikeRepo implements LikeRepo using a PostgreSQL database.
type PostgresLikeRepo struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewPostgresLikeRepo creates a new PostgresLikeRepo with the given database connection and logger.
func NewPostgresLikeRepo(db *sql.DB) *PostgresLikeRepo {
	return &PostgresLikeRepo{db: db, logger: logging.GetLogger()}
}

const (
	insertLikeQuery = `
		INSERT INTO likes (id, user_id, post_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, post_id, created_at, deleted_at
	`

	selectBaseQuery = `
		SELECT id, user_id, post_id, created_at, deleted_at
		FROM likes
		WHERE %s = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	selectByIDQuery = `
		SELECT id, user_id, post_id, created_at, deleted_at
		FROM likes
		WHERE id = $1 AND deleted_at IS NULL
	`

	softDeleteQuery = `
		UPDATE likes SET deleted_at = $1 WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	hardDeleteQuery = `
		DELETE FROM likes WHERE id = $1 AND user_id = $3
	`
)

// Create inserts a new like into the database.
// Returns an error if the like already exists or if user_id or post_id is empty.
func (r *PostgresLikeRepo) Create(ctx context.Context, s *model.Like) (*model.Like, error) {
	if s == nil {
		r.logger.Error("Create like failed: nil like")
		return nil, commonErrors.ErrInvalidArgument
	}
	if s.UserID == uuid.Nil {
		r.logger.Error("Create like failed: empty user ID")
		return nil, commonErrors.ErrInvalidArgument
	}
	if s.PostID == uuid.Nil {
		r.logger.Error("Create like failed: empty post ID")
		return nil, commonErrors.ErrInvalidArgument
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithError(err).Error("Create like failed: begin transaction")
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	newLike := &model.Like{
		ID:        s.ID,
		UserID:    s.UserID,
		PostID:    s.PostID,
		CreatedAt: now,
		DeletedAt: nil,
	}

	if newLike.ID == uuid.Nil {
		newLike.ID = uuid.New()
	}

	err = tx.QueryRowContext(ctx, insertLikeQuery,
		newLike.ID, newLike.UserID, newLike.PostID, newLike.CreatedAt).
		Scan(&newLike.ID, &newLike.UserID, &newLike.PostID, &newLike.CreatedAt, &newLike.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"user_id": s.UserID.String(),
				"post_id": s.PostID.String(),
			}).Error("Create like failed: duplicate like")
			return nil, commonErrors.ErrDuplicateLike
		}
		r.logger.WithError(err).Error("Create like failed")
		return nil, fmt.Errorf("create like: %w", err)
	}

	if err := tx.Commit(); err != nil {
		r.logger.WithError(err).Error("Create like failed: commit transaction")
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.WithField("id", newLike.ID.String()).Info("Like created")
	return newLike, nil
}

// GetByID retrieves a like by its ID.
// Returns an error if the ID is empty or the like is not found.
func (r *PostgresLikeRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Like, error) {
	if id == uuid.Nil {
		r.logger.Error("GetLikeByID failed: empty ID")
		return nil, commonErrors.ErrInvalidArgument
	}

	like := &model.Like{}
	err := r.db.QueryRowContext(ctx, selectByIDQuery, id).
		Scan(&like.ID, &like.UserID, &like.PostID, &like.CreatedAt, &like.DeletedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.WithField("id", id.String()).Error("GetLikeByID failed: like not found")
			return nil, commonErrors.ErrNotFound
		}
		r.logger.WithError(err).WithField("id", id.String()).Error("GetLikeByID failed")
		return nil, fmt.Errorf("get like by id: %w", err)
	}

	return like, nil
}

// GetByUserID retrieves all likes by a user with pagination.
// Returns an error if the user ID is empty or no likes are found.
func (r *PostgresLikeRepo) GetByUserID(ctx context.Context, id uuid.UUID, limit, offset int) ([]*model.Like, error) {
	if id == uuid.Nil {
		r.logger.Error("GetByUserID failed: empty user ID")
		return nil, commonErrors.ErrInvalidArgument
	}
	if limit <= 0 || offset < 0 {
		r.logger.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).Error("GetByUserID failed: invalid pagination parameters")
		return nil, commonErrors.ErrInvalidArgument
	}

	query := fmt.Sprintf(selectBaseQuery, "user_id")
	rows, err := r.db.QueryContext(ctx, query, id, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", id.String()).Error("GetByUserID failed")
		return nil, fmt.Errorf("get likes by user_id: %w", err)
	}
	defer rows.Close()

	var likes []*model.Like
	for rows.Next() {
		like := &model.Like{}
		if err := rows.Scan(&like.ID, &like.UserID, &like.PostID, &like.CreatedAt, &like.DeletedAt); err != nil {
			r.logger.WithError(err).Error("GetByUserID failed: scan like")
			return nil, fmt.Errorf("scan like: %w", err)
		}
		likes = append(likes, like)
	}

	if len(likes) == 0 {
		r.logger.WithField("user_id", id.String()).Error("GetByUserID failed: no likes found")
		return nil, commonErrors.ErrNotFound
	}

	return likes, nil
}

// GetByPostID retrieves all likes for a post with pagination.
// Returns an error if the post ID is empty or no likes are found.
func (r *PostgresLikeRepo) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*model.Like, error) {
	if postID == uuid.Nil {
		r.logger.Error("GetByPostID failed: empty post ID")
		return nil, commonErrors.ErrInvalidArgument
	}
	if limit <= 0 || offset < 0 {
		r.logger.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).Error("GetByPostID failed: invalid pagination parameters")
		return nil, commonErrors.ErrInvalidArgument
	}

	query := fmt.Sprintf(selectBaseQuery, "post_id")
	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		r.logger.WithError(err).WithField("post_id", postID.String()).Error("GetByPostID failed")
		return nil, fmt.Errorf("get likes by post_id: %w", err)
	}
	defer rows.Close()

	var likes []*model.Like
	for rows.Next() {
		like := &model.Like{}
		if err := rows.Scan(&like.ID, &like.UserID, &like.PostID, &like.CreatedAt, &like.DeletedAt); err != nil {
			r.logger.WithError(err).Error("GetByPostID failed: scan like")
			return nil, fmt.Errorf("scan like: %w", err)
		}
		likes = append(likes, like)
	}

	if len(likes) == 0 {
		r.logger.WithField("post_id", postID.String()).Error("GetByPostID failed: no likes found")
		return nil, commonErrors.ErrNotFound
	}

	return likes, nil
}

// Delete soft-deletes a like by its ID.
// Returns an error if the ID is empty or the like is not found.
func (r *PostgresLikeRepo) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if id == uuid.Nil {
		r.logger.Error("Unlike failed: empty ID")
		return commonErrors.ErrInvalidArgument
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithError(err).Error("Unlike failed: begin transaction")
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	result, err := tx.ExecContext(ctx, softDeleteQuery, now, id, userID)
	if err != nil {
		r.logger.WithError(err).WithField("id", id.String()).Error("Unlike failed")
		return fmt.Errorf("delete like: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("Unlike failed: check rows affected")
		return fmt.Errorf("delete like: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("id", id.String()).Error("Unlike failed: like not found")
		return commonErrors.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		r.logger.WithError(err).Error("Unlike failed: commit transaction")
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.WithField("id", id.String()).Info("Like deleted")
	return nil
}

// HardDelete permanently deletes a like by its ID.
// Returns an error if the ID is empty or the like is not found.
func (r *PostgresLikeRepo) HardDelete(ctx context.Context, id uuid.UUID, userId uuid.UUID) error {
	if id == uuid.Nil {
		r.logger.Error("HardDelete failed: empty ID")
		return commonErrors.ErrInvalidArgument
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithError(err).Error("HardDelete failed: begin transaction")
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, hardDeleteQuery, id, userId)
	if err != nil {
		r.logger.WithError(err).WithField("id", id.String()).Error("HardDelete failed")
		return fmt.Errorf("hard delete like: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).Error("HardDelete failed: check rows affected")
		return fmt.Errorf("hard delete like: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("id", id.String()).Error("HardDelete failed: like not found")
		return commonErrors.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		r.logger.WithError(err).Error("HardDelete failed: commit transaction")
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.logger.WithField("id", id.String()).Info("Like hard deleted")
	return nil
}
