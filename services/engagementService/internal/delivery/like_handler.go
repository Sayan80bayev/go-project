package delivery

import (
	"context"
	commonErrors "engagementService/internal/errors"
	"engagementService/internal/service"
	"engagementService/internal/transport/request"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

type LikeHandler struct {
	svc *service.LikeService
}

func NewLikeHandler(svc *service.LikeService) *LikeHandler {
	return &LikeHandler{svc: svc}
}

// GetLikeByID GET api/v1/likes/:id
func (h *LikeHandler) GetLikeByID(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	post, err := h.svc.GetByID(ctx, userId.(uuid.UUID))
	if err != nil {
		if errors.Is(err, commonErrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "like not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// Like POST api/v1/likes
func (h *LikeHandler) Like(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req := &request.LikeRequest{}
	err := c.ShouldBindBodyWithJSON(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = userID.(uuid.UUID)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	res, err := h.svc.Create(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetUserLikes GET api/v1/likes/user?limit=10&offset=0
func (h *LikeHandler) GetUserLikes(c *gin.Context) {

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := c.Query("limit")
	offset := c.Query("offset")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	res, err := h.svc.GetByUserID(ctx, userId.(uuid.UUID), limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetPostLikes GET api/v1/likes/post/:post_id?limit=10&offset=0
func (h *LikeHandler) GetPostLikes(c *gin.Context) {

	postId := c.Param("post_id")
	postUUID, err := uuid.Parse(postId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := c.Query("limit")
	offset := c.Query("offset")

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	res, err := h.svc.GetByUserID(ctx, postUUID, limitInt, offsetInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *LikeHandler) Unlike(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	likeId := c.Param("like_id")
	likeUUID, err := uuid.Parse(likeId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	err = h.svc.Delete(ctx, likeUUID, userId.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "like deleted"})
}
