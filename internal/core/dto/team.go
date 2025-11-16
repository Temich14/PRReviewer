package dto

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type GetTeamRequest struct {
	TeamName string `form:"TeamName" binding:"required"`
}

type DeactivateUsersRequest struct {
	TeamID  string   `json:"team_id" binding:"required"`
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
}

type DeactivateUsersResponse struct {
	DeactivatedUsers int `json:"deactivated_users"`
	ReassignedPRs    int `json:"reassigned_prs"`
	FailedReassigns  int `json:"failed_reassigns"`
}
