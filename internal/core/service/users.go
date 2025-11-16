package service

import (
	"PRReviewer/internal/core/entities"
	"context"
	"log/slog"
)

type UsersService struct {
	userRepo UserRepo
	tx       Transactor
	log      *slog.Logger
}

func NewUsersService(userRepo UserRepo, tx Transactor, log *slog.Logger) *UsersService {
	return &UsersService{userRepo: userRepo, tx: tx, log: log}
}

func (s *UsersService) SetIsActive(ctx context.Context, userID string, isActive bool) (*entities.User, error) {
	var user *entities.User
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		err := s.userRepo.SetIsActive(ctx, userID, isActive)
		if err != nil {
			s.log.Error("не удалось обновить статус пользователя", "error", err)
			return err
		}
		user, err = s.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			s.log.Error("не удалось получить пользователя", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		s.log.Error("транзакция завершилась с ошибкой", "error", err)
		return nil, err
	}
	return user, nil
}
