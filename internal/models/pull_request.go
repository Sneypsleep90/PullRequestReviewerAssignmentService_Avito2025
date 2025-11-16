package models

import "time"

type PullRequestStatus string

const (
	OPEN   PullRequestStatus = "OPEN"
	MERGED PullRequestStatus = "MERGED"
)

type PullRequest struct {
	PullRequestId     string            `db:"id" json:"pull_request_id" binding:"required"`
	PullRequestName   string            `db:"title" json:"pull_request_name" binding:"required"`
	AuthorId          string            `db:"author_id" json:"author_id" binding:"required"`
	Status            PullRequestStatus `db:"status" json:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers"`
	CreatedAt         *time.Time        `db:"created_at" json:"createdAt"`
	MergedAt          *time.Time        `db:"merged_at" json:"mergedAt"`
}

type PullRequestShort struct {
	PullRequestId   string            `db:"id" json:"pull_request_id"`
	PullRequestName string            `db:"title" json:"pull_request_name"`
	AuthorId        string            `db:"author_id" json:"author_id"`
	Status          PullRequestStatus `db:"status" json:"status"`
}
