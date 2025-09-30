package delivery

import (
	"context"
	commonErrors "engagementService/internal/errors"
	"engagementService/internal/service"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

type SubscriptionHandler struct {
	svc *service.SubscriptionService
}

func NewSubscriptionHandler(svc *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// Follow: POST /subscriptions/:followeeId/follow
func (h *SubscriptionHandler) Follow(c *gin.Context) {
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followeeIdStr := c.Param("followeeId")
	followeeID, err := uuid.Parse(followeeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid followee id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	err = h.svc.Follow(ctx, followerID.(uuid.UUID), followeeID)
	if err != nil {
		if errors.Is(err, commonErrors.ErrAlreadyFollowing) {
			c.JSON(http.StatusConflict, gin.H{"error": "already following"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "followed successfully"})
}

// Unfollow: DELETE /subscriptions/:followeeId/unfollow
func (h *SubscriptionHandler) Unfollow(c *gin.Context) {
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followeeIdStr := c.Param("followeeId")
	followeeID, err := uuid.Parse(followeeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid followee id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := h.svc.Unfollow(ctx, followerID.(uuid.UUID), followeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "unfollowed successfully"})
}

// GetFollowers: GET /subscriptions/:userId/followers?limit=10&offset=0
func (h *SubscriptionHandler) GetFollowers(c *gin.Context) {
	userIdStr := c.Param("userId")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	subs, err := h.svc.GetFollowers(ctx, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// GetFollowing: GET /subscriptions/:userId/following?limit=10&offset=0
func (h *SubscriptionHandler) GetFollowing(c *gin.Context) {
	userIdStr := c.Param("userId")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	subs, err := h.svc.GetFollowing(ctx, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// IsFollowing: GET /subscriptions/is-following/:followeeId
func (h *SubscriptionHandler) IsFollowing(c *gin.Context) {
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	followeeIdStr := c.Param("followeeId")
	followeeID, err := uuid.Parse(followeeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid followee id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	ok, err := h.svc.IsFollowing(ctx, followerID.(uuid.UUID), followeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_following": ok})
}

// Count: GET /subscriptions/:userId/count
func (h *SubscriptionHandler) Count(c *gin.Context) {
	userIdStr := c.Param("userId")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	followers, err := h.svc.CountFollowers(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	following, err := h.svc.CountFollowing(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
		"following": following,
	})
}
