package service

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/entities"
	"context"
)

type UserRepo interface {
	AddUsers(ctx context.Context, users []dto.TeamMember) error
	GetUserByID(ctx context.Context, userID string) (*entities.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	IsUserExist(ctx context.Context, userID string) (bool, error)
}
