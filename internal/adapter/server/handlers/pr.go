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

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, request dto.CreatePullRequest) (*entities.PullRequest, error)
	MergePullRequest(ctx context.Context, requestID string) (*entities.PullRequest, error)
	ReassignPullRequest(ctx context.Context, requestID string, oldUserID string) (*entities.PullRequest, error)
	GetUserReviewers(ctx context.Context, userID string) (*dto.GetPullRequestResponse, error)
}

func (h *PullRequestHandler) CreatePullRequest(c *gin.Context) {
	var req dto.CreatePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	pr, err := h.prSrv.CreatePullRequest(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
			return
		}
		if errors.Is(err, errs.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "PR_EXISTS", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pr)

}

func (h *PullRequestHandler) MergerPullRequest(c *gin.Context) {
	var req dto.MergePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	pr, err := h.prSrv.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, pr)
}

func (h *PullRequestHandler) ReassignPullRequest(c *gin.Context) {
	var req dto.ReassignReviewer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	pr, err := h.prSrv.ReassignPullRequest(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
			return
		}
		if errors.Is(err, errs.ErrAlreadyMerged) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "PR_MERGED", Message: err.Error()})
			return
		}
		if errors.Is(err, errs.ErrNoReviewersAvailable) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "NO_CANDIDATE", Message: err.Error()})
			return
		}

		if errors.Is(err, errs.ErrUserNotAssigned) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Code: "NOT_ASSIGNED", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pr)
}

func (h *PullRequestHandler) GetReview(c *gin.Context) {
	var req dto.UserIDQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prs, err := h.prSrv.GetUserReviewers(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, prs)
}
