package enums

type PRStatus string

const (
	PRStatusOpened PRStatus = "OPENED"
	PRStatusMerged PRStatus = "MERGED"
)

type Code string

const (
	CodeNotFound    Code = "NOT_FOUND"
	CodeMerged      Code = "PR_MERGED"
	CodeNoCandidate Code = "NO_CANDIDATE"
	CodeNotAssigned Code = "NOT_ASSIGNED"
	CodeTeamExists  Code = "TEAM_EXISTS"
)
