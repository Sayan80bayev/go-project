package service

import (
	"context"
	"engagementService/internal/model"
	"engagementService/internal/repository"
	"engagementService/internal/transport/request"
	"errors"
	"github.com/Sayan80bayev/go-project/pkg/logging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LikeService handles business logic for like-related operations.
type LikeService struct {
	repo   repository.LikeRepo
	logger *logrus.Logger
}

// NewLikeService creates a new LikeService with the given repository and logger.
func NewLikeService(repo repository.LikeRepo) *LikeService {
	return &LikeService{repo: repo, logger: logging.GetLogger()}
}

// Create adds a new like for a user and post.
// Returns an error if the request is nil or if user_id or post_id is empty.
func (s *LikeService) Create(ctx context.Context, r *request.LikeRequest) (*model.Like, error) {
	if r == nil {
		s.logger.Error("Create like failed: request is nil")
		return nil, errors.New("request cannot be nil")
	}
	if r.UserID == uuid.Nil {
		s.logger.Error("Create like failed: empty user ID")
		return nil, errors.New("user ID cannot be empty")
	}
	if r.PostID == uuid.Nil {
		s.logger.Error("Create like failed: empty post ID")
		return nil, errors.New("post ID cannot be empty")
	}

	like := &model.Like{
		PostID: r.PostID,
		UserID: r.UserID,
	}

	createdLike, err := s.repo.Create(ctx, like)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"user_id": r.UserID.String(),
			"post_id": r.PostID.String(),
		}).WithError(err).Error("Create like failed")
		return nil, err
	}
	s.logger.WithField("id", createdLike.ID.String()).Info("Like created")
	return createdLike, nil
}

// GetByID retrieves a like by its ID.
// Returns an error if the ID is empty or the like is not found.
func (s *LikeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Like, error) {
	if id == uuid.Nil {
		s.logger.Error("GetLikeByID failed: empty ID")
		return nil, errors.New("ID cannot be empty")
	}

	like, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithField("id", id.String()).WithError(err).Error("GetLikeByID failed")
		return nil, err
	}
	return like, nil
}

// GetByPostID retrieves all likes for a given post ID with pagination.
// Returns an error if the post ID is empty or no likes are found.
func (s *LikeService) GetByPostID(ctx context.Context, postID uuid.UUID, limit, offset int) ([]*model.Like, error) {
	if postID == uuid.Nil {
		s.logger.Error("GetByPostID failed: empty post ID")
		return nil, errors.New("post ID cannot be empty")
	}
	if limit <= 0 || offset < 0 {
		s.logger.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).Error("GetByPostID failed: invalid pagination parameters")
		return nil, errors.New("invalid pagination parameters")
	}

	likes, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		s.logger.WithField("post_id", postID.String()).WithError(err).Error("GetByPostID failed")
		return nil, err
	}
	return likes, nil
}

// GetByUserID retrieves all likes by a given user ID with pagination.
// Returns an error if the user ID is empty or no likes are found.
func (s *LikeService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Like, error) {
	if userID == uuid.Nil {
		s.logger.Error("GetByUserID failed: empty user ID")
		return nil, errors.New("user ID cannot be empty")
	}
	if limit <= 0 || offset < 0 {
		s.logger.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		}).Error("GetByUserID failed: invalid pagination parameters")
		return nil, errors.New("invalid pagination parameters")
	}

	likes, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.logger.WithField("user_id", userID.String()).WithError(err).Error("GetByUserID failed")
		return nil, err
	}
	return likes, nil
}

// Delete soft-deletes a like by its ID.
// Returns an error if the ID is empty or the like is not found.
func (s *LikeService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {

	if id == uuid.Nil || userID == uuid.Nil {
		s.logger.Error("Unlike failed: empty ID")
		return errors.New("ID cannot be empty")
	}

	if err := s.repo.Delete(ctx, id, userID); err != nil {
		s.logger.WithField("id", id.String()).WithError(err).Error("Unlike failed")
		return err
	}
	s.logger.WithField("id", id.String()).Info("Like deleted")
	return nil
}
