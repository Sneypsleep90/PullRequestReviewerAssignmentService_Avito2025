package integration

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/storage/postgres"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStorage_CreateUser(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewUserStorage(pool, logger)

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		user := &models.User{
			Id:       "user1",
			Username: "testuser",
			IsActive: true,
		}

		created, err := storage.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, user.Id, created.Id)
		assert.Equal(t, user.Username, created.Username)
		assert.Equal(t, user.IsActive, created.IsActive)
	})

	t.Run("duplicate id", func(t *testing.T) {
		user := &models.User{
			Id:       "user1",
			Username: "testuser2",
			IsActive: true,
		}

		_, err := storage.CreateUser(ctx, user)
		assert.Error(t, err)
	})
}

func TestUserStorage_GetUserByID(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewUserStorage(pool, logger)

	ctx := context.Background()

	// Create test user
	user := &models.User{
		Id:       "user1",
		Username: "testuser",
		IsActive: true,
	}
	_, err := storage.CreateUser(ctx, user)
	require.NoError(t, err)

	t.Run("user found", func(t *testing.T) {
		found, err := storage.GetUserByID(ctx, "user1")
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "user1", found.Id)
		assert.Equal(t, "testuser", found.Username)
	})

	t.Run("user not found", func(t *testing.T) {
		found, err := storage.GetUserByID(ctx, "nonexistent")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestUserStorage_GetUsersByTeam(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewUserStorage(pool, logger)

	ctx := context.Background()

	// Create team
	_, err := pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", "team1")
	require.NoError(t, err)

	// Create users
	user1 := &models.User{
		Id:       "user1",
		Username: "user1",
		IsActive: true,
		TeamName: "team1",
	}
	_, err = storage.CreateUser(ctx, user1)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "UPDATE users SET team_name = $1 WHERE id = $2", "team1", "user1")
	require.NoError(t, err)

	user2 := &models.User{
		Id:       "user2",
		Username: "user2",
		IsActive: true,
		TeamName: "team1",
	}
	_, err = storage.CreateUser(ctx, user2)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "UPDATE users SET team_name = $1 WHERE id = $2", "team1", "user2")
	require.NoError(t, err)

	t.Run("get users by team", func(t *testing.T) {
		users, err := storage.GetUsersByTeam(ctx, "team1")
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})
}
