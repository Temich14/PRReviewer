package repo

import (
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/errs"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

func (r *SQLRepo) CreateTeam(ctx context.Context, teamName string) (string, error) {
	teamID := uuid.New().String()
	query := `INSERT INTO teams (id, team_name) values ($1, $2)`

	executor := getExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query, teamID, teamName)
	if err != nil {
		return "", err
	}

	return teamID, nil
}

func (r *SQLRepo) IsTeamExistsByName(ctx context.Context, teamName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	var exists bool

	executor := getExecutor(ctx, r.db)
	err := executor.QueryRowContext(ctx, query, teamName).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

func (r *SQLRepo) AddMembersToTeam(ctx context.Context, teamID string, users []dto.TeamMember) error {
	if len(users) == 0 {
		return nil
	}

	var valueStrings []string
	var valueArgs []interface{}

	for i, user := range users {
		pos := i * 2
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", pos+1, pos+2))
		valueArgs = append(valueArgs, teamID, user.UserID)
	}

	query := fmt.Sprintf(
		"INSERT INTO team_members (team_id, user_id) VALUES %s ON CONFLICT (team_id, user_id) DO NOTHING",
		strings.Join(valueStrings, ", "),
	)

	executor := getExecutor(ctx, r.db)
	_, err := executor.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLRepo) GetTeamByName(ctx context.Context, teamName string) (*dto.Team, error) {

	query := `
				SELECT t.team_name, u.id, u.username, u.is_active 
				FROM teams t 
				LEFT JOIN team_members tm ON t.id = tm.team_id 
				LEFT JOIN users u ON tm.user_id = u.id 
				WHERE t.team_name = $1
			`

	executor := getExecutor(ctx, r.db)
	rows, err := executor.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var team *dto.Team
	members := make([]dto.TeamMember, 0)

	for rows.Next() {
		var teamName, userID, username string
		var isActive bool

		err := rows.Scan(&teamName, &userID, &username, &isActive)
		if err != nil {
			return nil, err
		}

		if team == nil {
			team = &dto.Team{
				TeamName: teamName,
				Members:  []dto.TeamMember{},
			}
		}

		if userID != "" {
			member := dto.TeamMember{
				UserID:   userID,
				Username: username,
				IsActive: isActive,
			}
			members = append(members, member)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if team == nil {
		return nil, errs.ErrNotFound
	}

	team.Members = members
	return team, nil
}
