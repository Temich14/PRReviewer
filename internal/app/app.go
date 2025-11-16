package app

import (
	"PRReviewer/config"
	"PRReviewer/internal/adapter/server"
	"PRReviewer/internal/adapter/server/handlers"
	"PRReviewer/internal/core/service"
	"PRReviewer/internal/infrastructure/data/repo"
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	server *server.Server
	log    *slog.Logger
}

func New(cfg *config.AppConfig) *App {

	db, err := sql.Open("pgx", cfg.DBCfg.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	repository := repo.New(db)

	transactor := repo.NewSQLTransactor(db)

	teamSrv := service.NewTeamService(repository, repository, transactor, logger)
	teamHnd := handlers.NewTeamHandler(teamSrv)

	userSrv := service.NewUsersService(repository, transactor, logger)
	userHnd := handlers.NewUsersHandler(userSrv)

	prSrv := service.NewPullRequestService(repository, repository, repository, transactor, logger)
	prHnd := handlers.NewPullRequestHandler(prSrv)

	httpServer := server.NewServer(cfg.ServerCfg, teamHnd, userHnd, prHnd)

	return &App{server: httpServer, log: logger}
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.stop(cancel)
	go a.server.Run()

	<-ctx.Done()

	a.log.Info("Shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	a.server.Stop(shutdownCtx)

	a.log.Info("shutdown complete")

}
func (a *App) stop(cancel context.CancelFunc) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	a.log.Info("signal to shutting down gracefully")
	cancel()
}
