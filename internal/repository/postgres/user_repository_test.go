package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"weather-api/internal/model"
	repo "weather-api/internal/repository/postgres"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	home, err := os.UserHomeDir()
	if err == nil {
		colimaSocket := filepath.Join(home, ".colima/default/docker.sock")
		if _, err := os.Stat(colimaSocket); err == nil && os.Getenv("DOCKER_HOST") == "" {
			os.Setenv("DOCKER_HOST", "unix://"+colimaSocket)
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
		}
	}

	// Locate the init.sql file relative to the test directory
	pwd, err := os.Getwd()
	require.NoError(t, err)
	initScript := filepath.Join(pwd, "../../../sql/001_init.sql")

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithInitScripts(initScript),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Open("postgres", connStr)
	require.NoError(t, err)

	require.NoError(t, db.Ping())

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repository := repo.NewUserRepository(db)

	userToCreate := &model.User{
		Email:        "test@example.com",
		PasswordHash: "hash123",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
	}

	// Test Create
	createdUser, err := repository.Create(context.Background(), userToCreate)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)
	assert.Equal(t, userToCreate.Email, createdUser.Email)
	assert.Equal(t, userToCreate.FirstName, createdUser.FirstName)

	// Test GetByID
	foundUser, err := repository.GetByID(context.Background(), createdUser.ID)
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, foundUser.ID)
	assert.Equal(t, createdUser.Email, foundUser.Email)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repository := repo.NewUserRepository(db)

	user, err := repository.GetByID(context.Background(), 999)
	require.ErrorIs(t, err, model.ErrUserNotFound)
	assert.Empty(t, user)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repository := repo.NewUserRepository(db)

	userToCreate := &model.User{
		Email:        "duplicate@example.com",
		PasswordHash: "hash",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
	}

	_, err := repository.Create(context.Background(), userToCreate)
	require.NoError(t, err)

	_, err = repository.Create(context.Background(), userToCreate)
	require.ErrorIs(t, err, model.ErrEmailAlreadyTaken)
}
