package service

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/errs"
	"context"
	"log/slog"
	"strings"
)

type TeamRepo interface {
	IsTeamExistsByName(ctx context.Context, teamName string) (bool, error)
	CreateTeam(ctx context.Context, teamName string) (string, error)
	AddMembersToTeam(ctx context.Context, teamID string, users []dto.TeamMember) error
	GetTeamByName(ctx context.Context, teamName string) (*dto.Team, error)
}

type TeamService struct {
	teamRepo TeamRepo
	UserRepo UserRepo
	tx       Transactor
	log      *slog.Logger
}

func NewTeamService(teamRepo TeamRepo, userRepo UserRepo, tx Transactor, log *slog.Logger) *TeamService {
	return &TeamService{teamRepo: teamRepo, UserRepo: userRepo, tx: tx, log: log}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *dto.Team) (*dto.Team, error) {
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		exists, err := s.teamRepo.IsTeamExistsByName(ctx, team.TeamName)
		if err != nil {
			return err
		}

		if exists {
			return errs.ErrAlreadyExists
		}

		id, err := s.teamRepo.CreateTeam(ctx, team.TeamName)
		if err != nil {
			return err
		}

		err = s.UserRepo.AddUsers(ctx, team.Members)
		if err != nil {
			return err
		}

		err = s.teamRepo.AddMembersToTeam(ctx, id, team.Members)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) GetTeamByName(ctx context.Context, teamName string) (*dto.Team, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, strings.TrimSpace(teamName))
	if err != nil {
		return nil, err
	}
	return team, nil
}
