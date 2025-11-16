package repo

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/entities"
	"PRReviewer/internal/core/errs"
	"context"
	"fmt"
	"strings"
)

func (r *SQLRepo) CreatePR(ctx context.Context, pr dto.CreatePullRequest) error {
	query := `INSERT INTO pull_requests (id, pr_name, author_id, status) VALUES ($1, $2, $3, $4)`

	executor := getExecutor(ctx, r.db)

	_, err := executor.ExecContext(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, "OPENED")
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLRepo) IsPRExists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM pull_requests WHERE id=$1)`
	executor := getExecutor(ctx, r.db)
	var exists bool
	err := executor.QueryRowContext(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *SQLRepo) AddReviewers(ctx context.Context, prID string, reviewers []string) error {
	if len(reviewers) == 0 {
		return nil
	}

	var valueStrings []string
	var valueArgs []interface{}

	for i, reviewer := range reviewers {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, prID, reviewer)
	}

	query := fmt.Sprintf(
		"INSERT INTO pull_request_reviewers (pr_id, reviewer_id) VALUES %s ON CONFLICT DO NOTHING",
		strings.Join(valueStrings, ","),
	)

	executor := getExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepo) GetPR(ctx context.Context, prID string) (*entities.PullRequest, error) {
	query := `
        SELECT p.id, p.pr_name, p.author_id, p.status, prr.reviewer_id, u.is_active
        FROM pull_requests p
        LEFT JOIN pull_request_reviewers prr ON p.id = prr.pr_id
        LEFT JOIN users u ON u.id = prr.reviewer_id
        WHERE p.id = $1
    `

	executor := getExecutor(ctx, r.db)
	rows, err := executor.QueryContext(ctx, query, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pr *entities.PullRequest
	reviewers := make([]dto.TeamMember, 0)

	for rows.Next() {
		var prID, prName, authorID, status, userID string
		var isActive bool

		err := rows.Scan(&prID, &prName, &authorID, &status, &userID, &isActive)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if pr == nil {
			pr = &entities.PullRequest{
				ID:        prID,
				Name:      prName,
				AuthorID:  authorID,
				Status:    status,
				Reviewers: reviewers,
			}
		}

		user := dto.TeamMember{
			IsActive: isActive,
			UserID:   userID,
		}

		reviewers = append(reviewers, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if pr == nil {
		return nil, errs.ErrNotFound
	}

	pr.Reviewers = reviewers
	return pr, nil
}

func (r *SQLRepo) MergePullRequest(ctx context.Context, requestID string) error {
	query := `UPDATE pull_requests SET status = 'MERGED' WHERE id = $1`

	executor := getExecutor(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, requestID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrNotFound
	}

	return nil
}

func (r *SQLRepo) ReassignPullRequest(ctx context.Context, prID string, oldReviewerID string, newReviewer string) error {
	query := `UPDATE pull_request_reviewers SET reviewer_id = $1 WHERE pr_id = $2 AND reviewer_id = $3`
	executor := getExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query, newReviewer, prID, oldReviewerID)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLRepo) GetUserPRReviews(ctx context.Context, userID string) ([]dto.PullRequestShort, error) {
	query := `
        SELECT 
            pr.id,
            pr.pr_name,
            pr.author_id,
            pr.status
        FROM pull_requests pr
        INNER JOIN pull_request_reviewers prr ON pr.id = prr.pr_id
        WHERE prr.reviewer_id = $1
    `

	executor := getExecutor(ctx, r.db)
	rows, err := executor.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []dto.PullRequestShort
	for rows.Next() {
		var pr dto.PullRequestShort
		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
