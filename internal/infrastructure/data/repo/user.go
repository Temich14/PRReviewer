package repo

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/entities"
	"PRReviewer/internal/core/errs"
	"context"
	"fmt"
	"strings"
)

func (r *SQLRepo) AddUsers(ctx context.Context, users []dto.TeamMember) error {
	if len(users) == 0 {
		return nil
	}

	var valueStrings []string
	var valueArgs []interface{}

	for i, user := range users {
		pos := i * 2
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", pos+1, pos+2))
		valueArgs = append(valueArgs, user.UserID, user.Username)
	}

	query := fmt.Sprintf(
		"INSERT INTO users (id, username) VALUES %s ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username",
		strings.Join(valueStrings, ", "),
	)

	executor := getExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepo) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE users SET is_active=$1 WHERE id=$2`

	executor := getExecutor(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, isActive, userID)
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

func (r *SQLRepo) IsUserExist(ctx context.Context, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`

	var exists bool

	executor := getExecutor(ctx, r.db)
	err := executor.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *SQLRepo) GetUserByID(ctx context.Context, userID string) (*entities.User, error) {
	query := `SELECT u.id, u.username, u.is_active, t.team_name FROM users u JOIN team_members tm ON u.id = tm.user_id JOIN teams t ON tm.team_id = t.id WHERE u.id=$1 LIMIT 1`
	var user entities.User
	var teamName *string

	executor := getExecutor(ctx, r.db)
	err := executor.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.IsActive,
		&teamName,
	)
	if err != nil {
		return nil, err
	}
	if teamName != nil {
		user.TeamName = *teamName
	}

	return &user, nil
}
