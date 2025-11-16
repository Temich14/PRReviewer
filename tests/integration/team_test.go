package integration

import (
	"PRReviewer/internal/adapter/server/handlers"
	"PRReviewer/internal/core/dto"
	"PRReviewer/internal/core/service"
	"PRReviewer/internal/infrastructure/data/repo"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/testcontainers/testcontainers-go"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type TeamIntegrationTestSuite struct {
	suite.Suite
	postgresContainer *postgres.PostgresContainer
	router            *gin.Engine
	teamHandler       *handlers.TeamHandler
	ctx               context.Context
}

func TestTeamIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TeamIntegrationTestSuite))
}

func (suite *TeamIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	postgresContainer, err := postgres.RunContainer(suite.ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	suite.Require().NoError(err)
	suite.postgresContainer = postgresContainer

	connStr, err := postgresContainer.ConnectionString(suite.ctx)
	suite.Require().NoError(err)

	db, err := sql.Open("pgx", connStr)
	suite.Require().NoError(err)

	repository := repo.New(db)
	transactor := repo.NewSQLTransactor(db)

	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	logger := slog.New(handler)

	teamService := service.NewTeamService(repository, repository, transactor, logger)

	suite.teamHandler = handlers.NewTeamHandler(teamService)

	suite.router = gin.Default()
	suite.router.POST("/team/add", suite.teamHandler.CreateTeam)

	time.Sleep(2 * time.Second)
}

func (suite *TeamIntegrationTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.Require().NoError(suite.postgresContainer.Terminate(suite.ctx))
	}
}

func (suite *TeamIntegrationTestSuite) SetupTest() {
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_Success() {
	teamRequest := dto.Team{
		TeamName: "payments",
		Members: []dto.TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
			{
				UserID:   "u2",
				Username: "Bob",
				IsActive: true,
			},
		},
	}

	body, err := json.Marshal(teamRequest)
	suite.NoError(err)

	req, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	suite.NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response struct {
		Team dto.Team `json:"team"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	assert.Equal(suite.T(), "payments", response.Team.TeamName)
	assert.Len(suite.T(), response.Team.Members, 2)
	assert.Equal(suite.T(), "u1", response.Team.Members[0].UserID)
	assert.Equal(suite.T(), "Alice", response.Team.Members[0].Username)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_DuplicateTeam() {
	teamRequest := dto.Team{
		TeamName: "backend",
		Members: []dto.TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	body, err := json.Marshal(teamRequest)
	suite.NoError(err)

	req, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	suite.NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	req2, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	suite.NoError(err)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusBadRequest, w2.Code)

	var errorResponse dto.ErrorResponse
	err = json.Unmarshal(w2.Body.Bytes(), &errorResponse)
	suite.NoError(err)

	assert.Equal(suite.T(), "TEAM_EXISTS", errorResponse.Code)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_InvalidJSON() {
	invalidJSON := `{"team_name": "payments", "members": [{"user_id": "u1"`

	req, err := http.NewRequest("POST", "/team/add", bytes.NewBufferString(invalidJSON))
	suite.NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_EmptyTeamName() {
	teamRequest := dto.Team{
		TeamName: "",
		Members: []dto.TeamMember{
			{
				UserID:   "u1",
				Username: "Alice",
				IsActive: true,
			},
		},
	}

	body, err := json.Marshal(teamRequest)
	suite.NoError(err)

	req, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body))
	suite.NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.True(suite.T(), w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_UpdateExistingUsers() {
	team1 := dto.Team{
		TeamName: "team1",
		Members: []dto.TeamMember{
			{
				UserID:   "u1",
				Username: "Alice_Old",
				IsActive: true,
			},
		},
	}

	body1, err := json.Marshal(team1)
	suite.NoError(err)

	req1, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body1))
	suite.NoError(err)
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)
	assert.Equal(suite.T(), http.StatusCreated, w1.Code)

	team2 := dto.Team{
		TeamName: "team2",
		Members: []dto.TeamMember{
			{
				UserID:   "u1",
				Username: "Alice_New",
				IsActive: true,
			},
		},
	}

	body2, err := json.Marshal(team2)
	suite.NoError(err)

	req2, err := http.NewRequest("POST", "/team/add", bytes.NewBuffer(body2))
	suite.NoError(err)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)
	assert.Equal(suite.T(), http.StatusCreated, w2.Code)

	var response2 struct {
		Team dto.Team `json:"team"`
	}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	suite.NoError(err)

	assert.Equal(suite.T(), "Alice_New", response2.Team.Members[0].Username)
}
