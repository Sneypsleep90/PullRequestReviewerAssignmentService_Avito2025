package integration

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/service/pull_request"
	"avito-autumn-2025/internal/storage/postgres"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestService_CreatePullRequest(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	prStorage := postgres.NewPullRequestStorage(pool, logger)
	userStorage := postgres.NewUserStorage(pool, logger)
	teamStorage := postgres.NewTeamStorage(pool, logger)

	service := pull_request.NewPullRequestService(&prStorage, &userStorage, &teamStorage, logger)

	ctx := context.Background()

	// Setup: create team and users
	_, err := pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"author1", "author1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"reviewer1", "reviewer1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"reviewer2", "reviewer2", true, "team1")
	require.NoError(t, err)

	t.Run("successful creation with reviewers", func(t *testing.T) {
		pr := &models.PullRequest{
			PullRequestId:   "pr1",
			PullRequestName: "Test PR",
			AuthorId:        "author1",
			Status:          models.OPEN,
		}

		created, err := service.CreatePullRequest(ctx, pr)
		require.NoError(t, err)
		assert.Equal(t, "pr1", created.PullRequestId)

		// Verify reviewers were assigned
		createdPR, err := prStorage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.LessOrEqual(t, len(createdPR.AssignedReviewers), 2)
		assert.NotContains(t, createdPR.AssignedReviewers, "author1") // Author should not be reviewer
	})

	t.Run("author not active", func(t *testing.T) {
		_, err := pool.Exec(ctx, "UPDATE users SET is_active = false WHERE id = $1", "author1")
		require.NoError(t, err)

		pr := &models.PullRequest{
			PullRequestId:   "pr2",
			PullRequestName: "Test PR 2",
			AuthorId:        "author1",
			Status:          models.OPEN,
		}

		_, err = service.CreatePullRequest(ctx, pr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})
}

func TestPullRequestService_MergePullRequest(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	prStorage := postgres.NewPullRequestStorage(pool, logger)
	userStorage := postgres.NewUserStorage(pool, logger)
	teamStorage := postgres.NewTeamStorage(pool, logger)

	service := pull_request.NewPullRequestService(&prStorage, &userStorage, &teamStorage, logger)

	ctx := context.Background()

	// Setup
	_, err := pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"author1", "author1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_requests (id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)",
		"pr1", "Test PR", "author1", "OPEN")
	require.NoError(t, err)

	t.Run("successful merge", func(t *testing.T) {
		err := service.MergePullRequest(ctx, "pr1")
		require.NoError(t, err)

		pr, err := prStorage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, models.MERGED, pr.Status)
	})

	t.Run("idempotent merge", func(t *testing.T) {
		err := service.MergePullRequest(ctx, "pr1")
		require.NoError(t, err) // Should not error on second merge
	})
}

func TestPullRequestService_ReassignReviewer(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	prStorage := postgres.NewPullRequestStorage(pool, logger)
	userStorage := postgres.NewUserStorage(pool, logger)
	teamStorage := postgres.NewTeamStorage(pool, logger)

	service := pull_request.NewPullRequestService(&prStorage, &userStorage, &teamStorage, logger)

	ctx := context.Background()

	// Setup
	_, err := pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"author1", "author1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"reviewer1", "reviewer1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"reviewer2", "reviewer2", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_requests (id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)",
		"pr1", "Test PR", "author1", "OPEN")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_request_reviewers (pr_id, user_id) VALUES ($1, $2)",
		"pr1", "reviewer1")
	require.NoError(t, err)

	t.Run("successful reassignment", func(t *testing.T) {
		err := service.ReassignReviewer(ctx, "pr1", "reviewer1")
		require.NoError(t, err)

		pr, err := prStorage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.NotContains(t, pr.AssignedReviewers, "reviewer1")
	})
}
