package handlers

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/errs"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team *dto.Team) (*dto.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (*dto.Team, error)
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	body := dto.Team{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	team, err := h.teamSrv.CreateTeam(c.Request.Context(), &body)
	if err != nil {
		if errors.Is(err, errs.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Code: "TEAM_EXISTS", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	var req dto.GetTeamRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	team, err := h.teamSrv.GetTeamByName(c.Request.Context(), req.TeamName)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Code: "NOT_FOUND", Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, team)
}
