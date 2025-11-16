package handlers

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PullRequestHandler struct {
	prService service.PullRequest
	log       logger.Logger
}

func NewPullRequestHandler(prService service.PullRequest, log logger.Logger) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
		log:       log,
	}
}

func (h *PullRequestHandler) PostPullRequestCreate(c *gin.Context) {
	h.log.Debug("Handler: Creating pull request request")

	var req models.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Handler: Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.PullRequestId == "" || req.PullRequestName == "" || req.AuthorId == "" {
		h.log.Error("Handler: Required fields missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pull_request_id, pull_request_name and author_id are required"})
		return
	}

	req.Status = models.OPEN
	now := time.Now()
	req.CreatedAt = &now

	pr, err := h.prService.CreatePullRequest(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Handler: Failed to create pull request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Handler: Pull request created successfully", "pr_id", pr.PullRequestId)
	c.JSON(http.StatusCreated, pr)
}

func (h *PullRequestHandler) PostPullRequestMerge(c *gin.Context) {
	h.log.Debug("Handler: Merging pull request request")

	var req struct {
		PullRequestId string `json:"pull_request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Handler: Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.PullRequestId == "" {
		h.log.Error("Handler: pull_request_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pull_request_id is required"})
		return
	}

	err := h.prService.MergePullRequest(c.Request.Context(), req.PullRequestId)
	if err != nil {
		h.log.Error("Handler: Failed to merge pull request", "error", err, "pr_id", req.PullRequestId)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Handler: Pull request merged successfully", "pr_id", req.PullRequestId)
	c.JSON(http.StatusOK, gin.H{"message": "Pull request merged successfully"})
}

func (h *PullRequestHandler) PostPullRequestReassign(c *gin.Context) {
	h.log.Debug("Handler: Reassigning reviewer request")

	var req struct {
		PullRequestId string `json:"pull_request_id" binding:"required"`
		OldUserId     string `json:"old_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Handler: Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.PullRequestId == "" || req.OldUserId == "" {
		h.log.Error("Handler: Required fields missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "pull_request_id and old_user_id are required"})
		return
	}

	err := h.prService.ReassignReviewer(c.Request.Context(), req.PullRequestId, req.OldUserId)
	if err != nil {
		h.log.Error("Handler: Failed to reassign reviewer", "error", err, "pr_id", req.PullRequestId)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Handler: Successfully reassigned reviewer", "pr_id", req.PullRequestId, "old_reviewer", req.OldUserId)
	c.JSON(http.StatusOK, gin.H{"message": "Reviewer reassigned successfully"})
}

func (h *PullRequestHandler) GetUsersGetReview(c *gin.Context) {
	h.log.Debug("Handler: Getting pull requests by reviewer request")

	reviewerID := c.Query("user_id")
	if reviewerID == "" {
		reviewerID = c.Query("reviewer_id")
	}
	if reviewerID == "" {
		h.log.Error("Handler: user_id or reviewer_id parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id or reviewer_id parameter is required"})
		return
	}

	h.log.Debug("Handler: Getting pull requests for reviewer", "reviewer_id", reviewerID)

	prs, err := h.prService.GetPullRequestsByReviewer(c.Request.Context(), reviewerID)
	if err != nil {
		h.log.Error("Handler: Failed to get pull requests by reviewer", "error", err, "reviewer_id", reviewerID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pull requests by reviewer"})
		return
	}

	h.log.Info("Handler: Successfully retrieved pull requests by reviewer", "reviewer_id", reviewerID, "count", len(prs))
	c.JSON(http.StatusOK, prs)
}

func (h *PullRequestHandler) GetReviewStatistics(c *gin.Context) {
	h.log.Debug("Handler: Getting review statistics request")

	stats, err := h.prService.GetReviewStatistics(c.Request.Context())
	if err != nil {
		h.log.Error("Handler: Failed to get review statistics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get review statistics"})
		return
	}

	h.log.Info("Handler: Review statistics retrieved successfully")
	c.JSON(http.StatusOK, stats)
}
