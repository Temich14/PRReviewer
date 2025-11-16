package service

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/entities"
	"PRReviewer/internal/core/errs"
	"context"
	"log/slog"
	"math/rand"
)

type PullRequestRepo interface {
	CreatePR(ctx context.Context, pr dto.CreatePullRequest) error
	AddReviewers(ctx context.Context, prID string, reviewers []string) error
	GetPR(ctx context.Context, prID string) (*entities.PullRequest, error)
	MergePullRequest(ctx context.Context, requestID string) error
	ReassignPullRequest(ctx context.Context, prID string, oldReviewerID string, newReviewer string) error
	GetUserPRReviews(ctx context.Context, userID string) ([]dto.PullRequestShort, error)
	IsPRExists(ctx context.Context, prID string) (bool, error)
}

type PullRequestService struct {
	prRepo   PullRequestRepo
	userRepo UserRepo
	TeamRepo TeamRepo
	tx       Transactor
	log      *slog.Logger
}

func NewPullRequestService(prRepo PullRequestRepo, userRepo UserRepo, TeamRepo TeamRepo, tx Transactor, log *slog.Logger) *PullRequestService {
	return &PullRequestService{prRepo, userRepo, TeamRepo, tx, log}
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, requestID string) (*entities.PullRequest, error) {
	err := s.prRepo.MergePullRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	pr, err := s.prRepo.GetPR(ctx, requestID)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr dto.CreatePullRequest) (*entities.PullRequest, error) {
	var pullRequest entities.PullRequest
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {

		exists, err := s.prRepo.IsPRExists(ctx, pr.PullRequestID)
		if err != nil {
			return err
		}
		if exists {
			return errs.ErrAlreadyExists
		}

		user, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
		if err != nil {
			return err
		}

		team, err := s.TeamRepo.GetTeamByName(ctx, user.TeamName)
		if err != nil {
			return err
		}

		reviewers, err := s.getReviewers(pr.AuthorID, team.Members, 2)
		if err != nil {
			return err
		}

		err = s.prRepo.CreatePR(ctx, pr)
		if err != nil {
			return err
		}

		err = s.prRepo.AddReviewers(ctx, pr.PullRequestID, reviewers)
		if err != nil {
			return err
		}

		getPR, err := s.prRepo.GetPR(ctx, pr.PullRequestID)
		if err != nil {
			return err
		}

		pullRequest = *getPR
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pullRequest, nil
}

func (s *PullRequestService) ReassignPullRequest(ctx context.Context, requestID string, oldUserID string) (*entities.PullRequest, error) {
	var pullRequest entities.PullRequest
	err := s.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		pr, err := s.prRepo.GetPR(ctx, requestID)
		if err != nil {
			return err
		}

		exists, err := s.userRepo.IsUserExist(ctx, oldUserID)
		if err != nil {
			return err
		}
		if !exists {
			return errs.ErrNotFound
		}

		if pr.Status == "MERGED" {
			return errs.ErrAlreadyMerged
		}

		ids := s.teamMembersToIDs(pr.Reviewers)

		author, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
		if err != nil {
			return err
		}

		if author.ID == oldUserID {
			return errs.ErrUserNotAssigned
		}

		team, err := s.TeamRepo.GetTeamByName(ctx, author.TeamName)
		if err != nil {
			return err
		}

		newReviewers, err := s.getReviewers(pr.AuthorID, team.Members, 1, ids...)
		if err != nil {
			return err
		}

		if len(newReviewers) < 1 {
			return errs.ErrNoReviewersAvailable
		}
		err = s.prRepo.ReassignPullRequest(ctx, requestID, oldUserID, newReviewers[0])
		if err != nil {
			return err
		}

		pr, err = s.prRepo.GetPR(ctx, requestID)
		if err != nil {
			return err
		}

		pullRequest = *pr

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pullRequest, nil
}

func (s *PullRequestService) GetUserReviewers(ctx context.Context, userID string) (*dto.GetPullRequestResponse, error) {

	exists, err := s.userRepo.IsUserExist(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errs.ErrNotFound
	}

	reviews, err := s.prRepo.GetUserPRReviews(ctx, userID)
	if err != nil {
		return nil, err
	}
	response := &dto.GetPullRequestResponse{
		UserID:       userID,
		PullRequests: reviews,
	}
	return response, nil
}

func (s *PullRequestService) getReviewers(authorID string, members []dto.TeamMember, limit int, excludeMembers ...string) ([]string, error) {
	reviewers := make([]dto.TeamMember, 0)
	excludeMap := make(map[string]bool)

	for _, excludedID := range excludeMembers {
		excludeMap[excludedID] = true
	}

	for _, member := range members {
		if member.UserID != authorID && member.IsActive && !excludeMap[member.UserID] {
			reviewers = append(reviewers, member)
		}
	}

	if len(reviewers) == 0 {
		return nil, errs.ErrNoReviewersAvailable
	}

	rand.Shuffle(len(reviewers), func(i, j int) {
		reviewers[i], reviewers[j] = reviewers[j], reviewers[i]
	})

	count := min(limit, len(reviewers))
	reviewerIDs := make([]string, count)
	for i, reviewer := range reviewers[:count] {
		reviewerIDs[i] = reviewer.UserID
	}

	return reviewerIDs, nil
}

func (s *PullRequestService) teamMembersToIDs(members []dto.TeamMember) []string {
	ids := make([]string, len(members))
	for i, member := range members {
		ids[i] = member.UserID
	}
	return ids
}
