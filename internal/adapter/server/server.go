package server

import (
	"PRReviewer/config"
	"PRReviewer/internal/adapter/server/handlers"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Server struct {
	server *http.Server
}

func NewServer(cfg *config.ServerConfig, teamHandler *handlers.TeamHandler, userHandler *handlers.UsersHandler, prHandler *handlers.PullRequestHandler) *Server {
	r := gin.New()
	api := r.Group("")
	teams := api.Group("/team")
	teams.POST("/add", teamHandler.CreateTeam)
	teams.GET("/get", teamHandler.GetTeam)

	users := api.Group("/users")
	users.POST("/setIsActive", userHandler.SetIsActive)
	users.GET("/getReview", prHandler.GetReview)

	pr := api.Group("/pullRequest")
	pr.POST("/create", prHandler.CreatePullRequest)
	pr.POST("/merge", prHandler.MergerPullRequest)
	pr.POST("/reassign", prHandler.ReassignPullRequest)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	return &Server{server: server}
}

func (s *Server) Run() {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server down: %s\n", err)
	}
}

func (s *Server) Stop(ctx context.Context) {
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}
}
