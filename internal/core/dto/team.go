package dto

type Team struct {
	TeamName string       `json:"team_name" binding:"required,min=1"`
	Members  []TeamMember `json:"members" binding:"required"`
}

type GetTeamRequest struct {
	TeamName string `form:"TeamName" binding:"required"`
}
