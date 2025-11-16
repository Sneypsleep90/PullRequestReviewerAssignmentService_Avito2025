package postgres

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestStorage struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewPullRequestStorage(db *pgxpool.Pool, log logger.Logger) PullRequestStorage {
	return PullRequestStorage{db: db, log: log}
}

func (p *PullRequestStorage) CreatePullRequest(ctx context.Context, pr *models.PullRequest, reviewers []string) error {
	p.log.Info("Creating pull request", "pr_id", pr.PullRequestId, "author_id", pr.AuthorId, "reviewers", reviewers)

	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.log.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO pull_requests (id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	_, err = tx.Exec(ctx, query, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, pr.Status, now)
	if err != nil {
		p.log.Error("Failed to insert pull request", "error", err, "pr_id", pr.PullRequestId)
		return fmt.Errorf("failed to insert pull request: %w", err)
	}

	for _, reviewerID := range reviewers {
		query = `INSERT INTO pull_request_reviewers (pr_id, user_id) VALUES ($1, $2)`
		_, err = tx.Exec(ctx, query, pr.PullRequestId, reviewerID)
		if err != nil {
			p.log.Error("Failed to insert reviewer", "error", err, "pr_id", pr.PullRequestId, "reviewer_id", reviewerID)
			return fmt.Errorf("failed to insert reviewer %s: %w", reviewerID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		p.log.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	p.log.Info("Successfully created pull request", "pr_id", pr.PullRequestId)
	return nil
}

func (p *PullRequestStorage) GetPullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	p.log.Debug("Getting pull request", "pr_id", prID)

	query := `
		SELECT id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	pr := &models.PullRequest{}
	var mergedAt sql.NullTime
	err := p.db.QueryRow(ctx, query, prID).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&pr.Status,
		&pr.CreatedAt,
		&mergedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.log.Warn("Pull request not found", "pr_id", prID)
			return nil, fmt.Errorf("pull request %s not found", prID)
		}
		p.log.Error("Failed to get pull request", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewersQuery := `SELECT user_id FROM pull_request_reviewers WHERE pr_id = $1`
	rows, err := p.db.Query(ctx, reviewersQuery, prID)
	if err != nil {
		p.log.Error("Failed to get reviewers", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			p.log.Error("Failed to scan reviewer", "error", err)
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	pr.AssignedReviewers = reviewers
	p.log.Debug("Successfully retrieved pull request", "pr_id", prID, "reviewers_count", len(reviewers))

	return pr, nil
}

func (p *PullRequestStorage) GetPullRequestsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequestShort, error) {
	p.log.Debug("Getting pull requests by reviewer", "reviewer_id", reviewerID)

	query := `
		SELECT pr.id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pull_request_reviewers prr ON pr.id = prr.pr_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := p.db.Query(ctx, query, reviewerID)
	if err != nil {
		p.log.Error("Failed to get pull requests by reviewer", "error", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to get pull requests by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequestShort
	for rows.Next() {
		pr := &models.PullRequestShort{}
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status); err != nil {
			p.log.Error("Failed to scan pull request", "error", err)
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	p.log.Debug("Successfully retrieved pull requests by reviewer", "reviewer_id", reviewerID, "count", len(prs))
	return prs, nil
}

func (p *PullRequestStorage) MergePullRequest(ctx context.Context, prID string) error {
	p.log.Info("Merging pull request", "pr_id", prID)

	var currentStatus models.PullRequestStatus
	checkQuery := `SELECT status FROM pull_requests WHERE id = $1`
	err := p.db.QueryRow(ctx, checkQuery, prID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.log.Warn("Pull request not found for merge", "pr_id", prID)
			return fmt.Errorf("pull request %s not found", prID)
		}
		p.log.Error("Failed to check pull request status", "error", err, "pr_id", prID)
		return fmt.Errorf("failed to check pull request status: %w", err)
	}

	if currentStatus == models.MERGED {
		p.log.Info("Pull request already merged", "pr_id", prID)
		return nil
	}

	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2 
		WHERE id = $3
	`
	_, err = p.db.Exec(ctx, query, models.MERGED, time.Now(), prID)
	if err != nil {
		p.log.Error("Failed to merge pull request", "error", err, "pr_id", prID)
		return fmt.Errorf("failed to merge pull request: %w", err)
	}

	p.log.Info("Successfully merged pull request", "pr_id", prID)
	return nil
}

func (p *PullRequestStorage) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) error {
	p.log.Info("Reassigning reviewer", "pr_id", prID, "old_reviewer", oldReviewerID)

	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.log.Error("Failed to begin transaction for reassignment", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var status models.PullRequestStatus
	checkQuery := `SELECT status FROM pull_requests WHERE id = $1`
	err = tx.QueryRow(ctx, checkQuery, prID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.log.Warn("Pull request not found for reassignment", "pr_id", prID)
			return fmt.Errorf("pull request %s not found", prID)
		}
		p.log.Error("Failed to check pull request status for reassignment", "error", err, "pr_id", prID)
		return fmt.Errorf("failed to check pull request status: %w", err)
	}

	if status == models.MERGED {
		p.log.Warn("Cannot reassign reviewers on merged pull request", "pr_id", prID)
		return fmt.Errorf("cannot reassign reviewers on merged pull request %s", prID)
	}

	var exists bool
	existsQuery := `SELECT EXISTS(SELECT 1 FROM pull_request_reviewers WHERE pr_id = $1 AND user_id = $2)`
	err = tx.QueryRow(ctx, existsQuery, prID, oldReviewerID).Scan(&exists)
	if err != nil {
		p.log.Error("Failed to check reviewer existence", "error", err, "pr_id", prID, "reviewer_id", oldReviewerID)
		return fmt.Errorf("failed to check reviewer existence: %w", err)
	}

	if !exists {
		p.log.Warn("Old reviewer not assigned to pull request", "pr_id", prID, "old_reviewer", oldReviewerID)
		return fmt.Errorf("reviewer %s is not assigned to pull request %s", oldReviewerID, prID)
	}

	var teamName string
	teamQuery := `SELECT team_name FROM users WHERE id = $1`
	err = tx.QueryRow(ctx, teamQuery, oldReviewerID).Scan(&teamName)
	if err != nil {
		p.log.Error("Failed to get reviewer team", "error", err, "reviewer_id", oldReviewerID)
		return fmt.Errorf("failed to get reviewer team: %w", err)
	}

	var newReviewerID string
	newReviewerQuery := `
		SELECT id FROM users 
		WHERE team_name = $1 
		AND is_active = true 
		AND id != $2
		AND id NOT IN (
			SELECT user_id FROM pull_request_reviewers WHERE pr_id = $3
		)
		ORDER BY RANDOM()
		LIMIT 1
	`
	err = tx.QueryRow(ctx, newReviewerQuery, teamName, oldReviewerID, prID).Scan(&newReviewerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.log.Warn("No available reviewers found for reassignment", "pr_id", prID, "team_name", teamName)
			return fmt.Errorf("no available reviewers found in team %s", teamName)
		}
		p.log.Error("Failed to find new reviewer", "error", err)
		return fmt.Errorf("failed to find new reviewer: %w", err)
	}

	updateQuery := `
		UPDATE pull_request_reviewers 
		SET user_id = $1 
		WHERE pr_id = $2 AND user_id = $3
	`
	_, err = tx.Exec(ctx, updateQuery, newReviewerID, prID, oldReviewerID)
	if err != nil {
		p.log.Error("Failed to reassign reviewer", "error", err, "pr_id", prID, "old_reviewer", oldReviewerID, "new_reviewer", newReviewerID)
		return fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		p.log.Error("Failed to commit reassignment transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	p.log.Info("Successfully reassigned reviewer", "pr_id", prID, "old_reviewer", oldReviewerID, "new_reviewer", newReviewerID)
	return nil
}

func (p *PullRequestStorage) GetReviewStatistics(ctx context.Context) (*models.ReviewStatistics, error) {
	p.log.Debug("Getting review statistics")

	stats := &models.ReviewStatistics{
		GeneratedAt: time.Now(),
	}

	var totalPRs int
	err := p.db.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests`).Scan(&totalPRs)
	if err != nil {
		p.log.Error("Failed to get total PRs count", "error", err)
		return nil, fmt.Errorf("failed to get total PRs count: %w", err)
	}
	stats.TotalPRs = totalPRs

	var openPRs int
	err = p.db.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'`).Scan(&openPRs)
	if err != nil {
		p.log.Error("Failed to get open PRs count", "error", err)
		return nil, fmt.Errorf("failed to get open PRs count: %w", err)
	}
	stats.OpenPRs = openPRs

	var mergedPRs int
	err = p.db.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'`).Scan(&mergedPRs)
	if err != nil {
		p.log.Error("Failed to get merged PRs count", "error", err)
		return nil, fmt.Errorf("failed to get merged PRs count: %w", err)
	}
	stats.MergedPRs = mergedPRs

	reviewerRows, err := p.db.Query(ctx, `
		SELECT u.id, u.username, COUNT(prr.pr_id) as assigned_count, MAX(pr.created_at) as last_assigned
		FROM users u
		LEFT JOIN pull_request_reviewers prr ON u.id = prr.user_id
		LEFT JOIN pull_requests pr ON prr.pr_id = pr.id
		WHERE u.is_active = true
		GROUP BY u.id, u.username
		ORDER BY assigned_count DESC
	`)
	if err != nil {
		p.log.Error("Failed to get reviewer statistics", "error", err)
		return nil, fmt.Errorf("failed to get reviewer statistics: %w", err)
	}
	defer reviewerRows.Close()

	for reviewerRows.Next() {
		var reviewerStats models.ReviewerStatistics
		var lastAssigned *time.Time

		err := reviewerRows.Scan(
			&reviewerStats.ReviewerID,
			&reviewerStats.ReviewerName,
			&reviewerStats.AssignedPRsCount,
			&lastAssigned,
		)
		if err != nil {
			p.log.Error("Failed to scan reviewer statistics row", "error", err)
			return nil, fmt.Errorf("failed to scan reviewer statistics row: %w", err)
		}

		if lastAssigned != nil {
			reviewerStats.LastAssignedAt = lastAssigned
		}

		stats.ReviewerStats = append(stats.ReviewerStats, reviewerStats)
	}

	teamRows, err := p.db.Query(ctx, `
		SELECT 
			t.name as team_name,
			COUNT(u.id) as member_count,
			COUNT(CASE WHEN u.is_active = true THEN 1 END) as active_member_count,
			COUNT(pr.id) as prs_created
		FROM teams t
		LEFT JOIN users u ON t.name = u.team_name
		LEFT JOIN pull_requests pr ON u.id = pr.author_id
		GROUP BY t.name
		ORDER BY team_name
	`)
	if err != nil {
		p.log.Error("Failed to get team statistics", "error", err)
		return nil, fmt.Errorf("failed to get team statistics: %w", err)
	}
	defer teamRows.Close()

	for teamRows.Next() {
		var teamStats models.TeamStatistics

		err := teamRows.Scan(
			&teamStats.TeamName,
			&teamStats.MemberCount,
			&teamStats.ActiveMemberCount,
			&teamStats.PRsCreated,
		)
		if err != nil {
			p.log.Error("Failed to scan team statistics row", "error", err)
			return nil, fmt.Errorf("failed to scan team statistics row: %w", err)
		}

		stats.TeamStats = append(stats.TeamStats, teamStats)
	}

	p.log.Info("Review statistics retrieved successfully",
		"total_prs", stats.TotalPRs,
		"open_prs", stats.OpenPRs,
		"merged_prs", stats.MergedPRs,
		"reviewer_count", len(stats.ReviewerStats),
		"team_count", len(stats.TeamStats))

	return stats, nil
}

func (p *PullRequestStorage) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUser string) ([]*models.User, error) {
	p.log.Debug("Getting active team members", "team_name", teamName, "exclude_user", excludeUser)

	query := `
		SELECT id, username, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND id != $2
		ORDER BY username
	`

	rows, err := p.db.Query(ctx, query, teamName, excludeUser)
	if err != nil {
		p.log.Error("Failed to get active team members", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to get active team members: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.IsActive); err != nil {
			p.log.Error("Failed to scan user", "error", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	p.log.Debug("Successfully retrieved active team members", "team_name", teamName, "count", len(users))
	return users, nil
}
