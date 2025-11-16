package integration

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/storage/postgres"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestStorage_CreatePullRequest(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewPullRequestStorage(pool, logger)

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

	t.Run("successful creation", func(t *testing.T) {
		pr := &models.PullRequest{
			PullRequestId:   "pr1",
			PullRequestName: "Test PR",
			AuthorId:        "author1",
			Status:          models.OPEN,
		}
		reviewers := []string{"reviewer1", "reviewer2"}

		err := storage.CreatePullRequest(ctx, pr, reviewers)
		require.NoError(t, err)

		// Verify PR was created
		created, err := storage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, "pr1", created.PullRequestId)
		assert.Equal(t, "Test PR", created.PullRequestName)
		assert.Len(t, created.AssignedReviewers, 2)
	})
}

func TestPullRequestStorage_MergePullRequest(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewPullRequestStorage(pool, logger)

	ctx := context.Background()

	// Setup: create PR
	_, err := pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO users (id, username, is_active, team_name) VALUES ($1, $2, $3, $4)",
		"author1", "author1", true, "team1")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_requests (id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)",
		"pr1", "Test PR", "author1", "OPEN")
	require.NoError(t, err)

	t.Run("successful merge", func(t *testing.T) {
		err := storage.MergePullRequest(ctx, "pr1")
		require.NoError(t, err)

		pr, err := storage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, models.MERGED, pr.Status)
		assert.NotNil(t, pr.MergedAt)
	})

	t.Run("idempotent merge", func(t *testing.T) {
		err := storage.MergePullRequest(ctx, "pr1")
		require.NoError(t, err)

		pr, err := storage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.Equal(t, models.MERGED, pr.Status)
	})
}

func TestPullRequestStorage_ReassignReviewer(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewPullRequestStorage(pool, logger)

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

	_, err = pool.Exec(ctx, "INSERT INTO pull_requests (id, pull_request_name, author_id, status) VALUES ($1, $2, $3, $4)",
		"pr1", "Test PR", "author1", "OPEN")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_request_reviewers (pr_id, user_id) VALUES ($1, $2)",
		"pr1", "reviewer1")
	require.NoError(t, err)

	t.Run("successful reassignment", func(t *testing.T) {
		err := storage.ReassignReviewer(ctx, "pr1", "reviewer1")
		require.NoError(t, err)

		pr, err := storage.GetPullRequest(ctx, "pr1")
		require.NoError(t, err)
		assert.Contains(t, pr.AssignedReviewers, "reviewer2")
		assert.NotContains(t, pr.AssignedReviewers, "reviewer1")
	})

	t.Run("cannot reassign merged PR", func(t *testing.T) {
		// Merge PR
		err := storage.MergePullRequest(ctx, "pr1")
		require.NoError(t, err)

		// Try to reassign
		err = storage.ReassignReviewer(ctx, "pr1", "reviewer2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merged")
	})
}

func TestPullRequestStorage_GetPullRequestsByReviewer(t *testing.T) {
	pool, cleanup := SetupTestDB(t)
	defer cleanup()

	logger := logger.NewStdLogger()
	storage := postgres.NewPullRequestStorage(pool, logger)

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

	_, err = pool.Exec(ctx, "INSERT INTO pull_requests (id, pull_request_name, author_id, status, created_at) VALUES ($1, $2, $3, $4, $5)",
		"pr1", "PR 1", "author1", "OPEN", time.Now())
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO pull_request_reviewers (pr_id, user_id) VALUES ($1, $2)",
		"pr1", "reviewer1")
	require.NoError(t, err)

	t.Run("get PRs by reviewer", func(t *testing.T) {
		prs, err := storage.GetPullRequestsByReviewer(ctx, "reviewer1")
		require.NoError(t, err)
		assert.Len(t, prs, 1)
		assert.Equal(t, "pr1", prs[0].PullRequestId)
	})
}
