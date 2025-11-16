package entities

import "PRReviewer/internal/core/dto"

type PullRequest struct {
	ID        string           `json:"pull_request_id"`
	Name      string           `json:"pull_request_name"`
	AuthorID  string           `json:"author_id"`
	Status    string           `json:"status"`
	Reviewers []dto.TeamMember `json:"assigned_reviewers"`
}

type PullRequestResponse struct {
	ID        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorID  string   `json:"author_id"`
	Status    string   `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

type ReassignedPullRequest struct {
	PR         PullRequest `json:"pr"`
	ReplacedBY string      `json:"replaced_by"`
}
