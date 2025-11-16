package storage

import (
	"avito-autumn-2025/internal/models"
	"context"
)

type User interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUsersByTeam(ctx context.Context, teamName string) ([]*models.User, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) error
}

type Team interface {
	CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error)
	GetTeamWithMembers(ctx context.Context, teamName string) (*models.Team, error)
}

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest, reviewers []string) error
	GetPullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	GetPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error)
	MergePullRequest(ctx context.Context, prID string) error
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) error
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUser string) ([]*models.User, error)
	GetReviewStatistics(ctx context.Context) (*models.ReviewStatistics, error)
}
