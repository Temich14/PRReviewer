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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type TeamIntegrationTestSuite struct {
	suite.Suite
	postgresContainer *postgres.PostgresContainer
	router            *gin.Engine
	teamHandler       *handlers.TeamHandler
	ctx               context.Context
	db                *sql.DB
}

func TestTeamIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TeamIntegrationTestSuite))
}

func (suite *TeamIntegrationTestSuite) migrate(db *sql.DB, relativePath string) error {

	content, err := os.ReadFile("../../migrations/1_init.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	suite.T().Logf("Migration applied successfully: %s", relativePath)
	return nil
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
				WithStartupTimeout(30*time.Second)),
	)
	suite.Require().NoError(err)
	suite.postgresContainer = postgresContainer

	connStr, err := postgresContainer.ConnectionString(suite.ctx)
	suite.Require().NoError(err)

	db, err := sql.Open("pgx", connStr)
	suite.Require().NoError(err)

	suite.db = db

	err = db.PingContext(suite.ctx)
	suite.Require().NoError(err, "Failed to connect to database")

	err = suite.migrate(db, "migrations/1_init.up.sql")
	suite.Require().NoError(err, "Failed to apply migrations")

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

	suite.T().Logf("SetupSuite completed successfully")
}

func (suite *TeamIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.postgresContainer != nil {
		suite.Require().NoError(suite.postgresContainer.Terminate(suite.ctx))
	}
}

func (suite *TeamIntegrationTestSuite) SetupTest() {
	if suite.db != nil {
		_, err := suite.db.ExecContext(suite.ctx, `
            DELETE FROM team_members;
            DELETE FROM teams;
            DELETE FROM users;
        `)
		if err != nil {
			suite.T().Logf("error: %s", err)
		}
	}
}

func (suite *TeamIntegrationTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		suite.Require().NoError(err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

func (suite *TeamIntegrationTestSuite) parseTeamResponse(w *httptest.ResponseRecorder) dto.Team {
	var res dto.Team
	err := json.Unmarshal(w.Body.Bytes(), &res)
	suite.Require().NoError(err)
	return res
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_WhenValidRequest_ShouldCreateTeamWithMembers() {
	// Arrange
	teamRequest := dto.Team{
		TeamName: "payments",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	// Act
	response := suite.makeRequest("POST", "/team/add", teamRequest)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, response.Code)

	createdTeam := suite.parseTeamResponse(response)
	assert.Equal(suite.T(), "payments", createdTeam.TeamName)
	assert.Len(suite.T(), createdTeam.Members, 2)
	assert.Equal(suite.T(), "u1", createdTeam.Members[0].UserID)
	assert.Equal(suite.T(), "Alice", createdTeam.Members[0].Username)
	assert.Equal(suite.T(), "u2", createdTeam.Members[1].UserID)
	assert.Equal(suite.T(), "Bob", createdTeam.Members[1].Username)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_WhenDuplicateTeam_ShouldReturnConflictError() {
	// Arrange
	teamRequest := dto.Team{
		TeamName: "backend",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	// Act
	firstResponse := suite.makeRequest("POST", "/team/add", teamRequest)
	assert.Equal(suite.T(), http.StatusCreated, firstResponse.Code)

	secondResponse := suite.makeRequest("POST", "/team/add", teamRequest)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, secondResponse.Code)

	var errorResponse dto.ErrorResponse
	err := json.Unmarshal(secondResponse.Body.Bytes(), &errorResponse)
	suite.Require().NoError(err)
	assert.Equal(suite.T(), "TEAM_EXISTS", errorResponse.Code)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_WhenInvalidJSON_ShouldReturnBadRequest() {
	// Arrange
	invalidJSON := `{"team_name": "payments", "members": [{"user_id": "u1"`

	// Act
	req, _ := http.NewRequest("POST", "/team/add", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_WhenEmptyTeamName_ShouldReturnBadRequest() {
	// Arrange
	teamRequest := dto.Team{
		TeamName: "",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	// Act
	response := suite.makeRequest("POST", "/team/add", teamRequest)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, response.Code)
}

func (suite *TeamIntegrationTestSuite) TestCreateTeam_WhenUserExistsInOtherTeam_ShouldUpdateUser() {
	// Arrange
	firstTeam := dto.Team{
		TeamName: "team1",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Alice_Old", IsActive: true},
		},
	}

	secondTeam := dto.Team{
		TeamName: "team2",
		Members: []dto.TeamMember{
			{UserID: "u1", Username: "Alice_New", IsActive: true},
		},
	}

	// Act
	firstResponse := suite.makeRequest("POST", "/team/add", firstTeam)
	assert.Equal(suite.T(), http.StatusCreated, firstResponse.Code)

	secondResponse := suite.makeRequest("POST", "/team/add", secondTeam)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, secondResponse.Code)

	updatedTeam := suite.parseTeamResponse(secondResponse)
	assert.Equal(suite.T(), "Alice_New", updatedTeam.Members[0].Username)
}
