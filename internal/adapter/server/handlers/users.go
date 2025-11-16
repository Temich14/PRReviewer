package handlers

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/entities"
	"PRReviewer/internal/core/errs"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UsersService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entities.User, error)
}

func (h *UsersHandler) SetIsActive(c *gin.Context) {
	var req dto.SetUserActiveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.userSrv.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
