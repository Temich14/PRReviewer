package handlers

type TeamHandler struct {
	teamSrv TeamService
}

func NewTeamHandler(teamSrv TeamService) *TeamHandler {
	return &TeamHandler{teamSrv: teamSrv}
}

type UsersHandler struct {
	userSrv UsersService
}

func NewUsersHandler(userSrv UsersService) *UsersHandler {
	return &UsersHandler{userSrv: userSrv}
}

type PullRequestHandler struct {
	prSrv PullRequestService
}

func NewPullRequestHandler(prSrv PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prSrv}
}
