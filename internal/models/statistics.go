package models

import "time"

type ReviewStatistics struct {
	TotalPRs      int                  `json:"total_prs"`
	OpenPRs       int                  `json:"open_prs"`
	MergedPRs     int                  `json:"merged_prs"`
	ReviewerStats []ReviewerStatistics `json:"reviewer_stats"`
	TeamStats     []TeamStatistics     `json:"team_stats"`
	GeneratedAt   time.Time            `json:"generated_at"`
}

type ReviewerStatistics struct {
	ReviewerID       string     `json:"reviewer_id"`
	ReviewerName     string     `json:"reviewer_name"`
	AssignedPRsCount int        `json:"assigned_prs_count"`
	LastAssignedAt   *time.Time `json:"last_assigned_at,omitempty"`
}

type TeamStatistics struct {
	TeamName          string `json:"team_name"`
	MemberCount       int    `json:"member_count"`
	ActiveMemberCount int    `json:"active_member_count"`
	PRsCreated        int    `json:"prs_created"`
}
