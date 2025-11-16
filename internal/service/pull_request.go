package service

import (
	"avito-autumn-2025/internal/models"
	"context"
)

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) error
	ReassignReviewer(ctx context.Context, prID, oldUserID string) error
	GetPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error)
	GetReviewStatistics(ctx context.Context) (*models.ReviewStatistics, error)
}
