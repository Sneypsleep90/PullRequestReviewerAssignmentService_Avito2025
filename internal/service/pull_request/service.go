package pull_request

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/storage"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type PullRequestService struct {
	prStorage   storage.PullRequest
	userStorage storage.User
	teamStorage storage.Team
	log         logger.Logger
}

func NewPullRequestService(
	prStorage storage.PullRequest,
	userStorage storage.User,
	teamStorage storage.Team,
	log logger.Logger,
) *PullRequestService {
	return &PullRequestService{
		prStorage:   prStorage,
		userStorage: userStorage,
		teamStorage: teamStorage,
		log:         log,
	}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	s.log.Info("Creating pull request", "pr_id", pr.PullRequestId, "author_id", pr.AuthorId)

	author, err := s.userStorage.GetUserByID(ctx, pr.AuthorId)
	if err != nil {
		s.log.Error("Failed to get author", "error", err, "author_id", pr.AuthorId)
		return nil, fmt.Errorf("failed to get author: %w", err)
	}
	if author == nil {
		s.log.Warn("Author not found", "author_id", pr.AuthorId)
		return nil, fmt.Errorf("author %s not found", pr.AuthorId)
	}

	if !author.IsActive {
		s.log.Warn("Author is not active", "author_id", pr.AuthorId)
		return nil, fmt.Errorf("author %s is not active", pr.AuthorId)
	}

	teamMembers, err := s.prStorage.GetActiveTeamMembers(ctx, author.TeamName, author.Id)
	if err != nil {
		s.log.Error("Failed to get team members", "error", err, "team_name", author.TeamName)
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	reviewers := s.selectReviewers(teamMembers, 2)

	err = s.prStorage.CreatePullRequest(ctx, pr, reviewers)
	if err != nil {
		s.log.Error("Failed to create pull request", "error", err, "pr_id", pr.PullRequestId)
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	pr.AssignedReviewers = reviewers

	s.log.Info("Successfully created pull request", "pr_id", pr.PullRequestId, "reviewers_count", len(reviewers))
	return pr, nil
}

func (s *PullRequestService) MergePullRequest(ctx context.Context, prID string) error {
	s.log.Info("Merging pull request", "pr_id", prID)

	err := s.prStorage.MergePullRequest(ctx, prID)
	if err != nil {
		s.log.Error("Failed to merge pull request", "error", err, "pr_id", prID)
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	s.log.Info("Successfully merged pull request", "pr_id", prID)
	return nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) error {
	s.log.Info("Reassigning reviewer", "pr_id", prID, "old_reviewer", oldUserID)

	oldReviewer, err := s.userStorage.GetUserByID(ctx, oldUserID)
	if err != nil {
		s.log.Error("Failed to get old reviewer", "error", err, "reviewer_id", oldUserID)
		return fmt.Errorf("failed to get old reviewer: %w", err)
	}
	if oldReviewer == nil {
		s.log.Warn("Old reviewer not found", "reviewer_id", oldUserID)
		return fmt.Errorf("reviewer %s not found", oldUserID)
	}

	err = s.prStorage.ReassignReviewer(ctx, prID, oldUserID)
	if err != nil {
		s.log.Error("Failed to reassign reviewer", "error", err, "pr_id", prID)
		return fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	s.log.Info("Successfully reassigned reviewer", "pr_id", prID, "old_reviewer", oldUserID)
	return nil
}

func (s *PullRequestService) GetReviewStatistics(ctx context.Context) (*models.ReviewStatistics, error) {
	s.log.Debug("Service: Getting review statistics")

	stats, err := s.prStorage.GetReviewStatistics(ctx)
	if err != nil {
		s.log.Error("Service: Failed to get review statistics", "error", err)
		return nil, fmt.Errorf("failed to get review statistics: %w", err)
	}

	s.log.Info("Service: Review statistics retrieved successfully")
	return stats, nil
}

func (s *PullRequestService) GetPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error) {
	s.log.Debug("Getting pull requests by reviewer", "reviewer_id", reviewerID)

	prs, err := s.prStorage.GetPullRequestsByReviewer(ctx, reviewerID)
	if err != nil {
		s.log.Error("Failed to get pull requests by reviewer", "error", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to get pull requests by reviewer: %w", err)
	}

	s.log.Debug("Successfully retrieved pull requests by reviewer", "reviewer_id", reviewerID, "count", len(prs))
	return prs, nil
}

func (s *PullRequestService) selectReviewers(members []*models.User, maxCount int) []string {
	if len(members) == 0 {
		return []string{}
	}

	rand.Seed(time.Now().UnixNano())
	for i := range members {
		j := rand.Intn(i + 1)
		members[i], members[j] = members[j], members[i]
	}

	count := len(members)
	if count > maxCount {
		count = maxCount
	}

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = members[i].Id
	}

	return reviewers
}
