package handlers

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamService service.Team
	log         logger.Logger
}

func NewTeamHandler(teamService service.Team, log logger.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		log:         log,
	}
}

func (h *TeamHandler) PostTeamAdd(c *gin.Context) {
	h.log.Debug("Handler: Creating team request")

	var req models.Team
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Handler: Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.Name == "" {
		h.log.Error("Handler: Team name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "team_name is required"})
		return
	}

	team, err := h.teamService.CreateTeam(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Handler: Failed to create team", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Handler: Team created successfully", "team_name", team.Name)
	c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) GetTeamTeamName(c *gin.Context) {
	h.log.Debug("Handler: Getting team by name request")

	teamName := c.Param("teamName")
	if teamName == "" {
		h.log.Error("Handler: team name parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team name parameter is required"})
		return
	}

	h.log.Debug("Handler: Getting team", "team_name", teamName)

	team, err := h.teamService.GetTeamWithMembers(c.Request.Context(), teamName)
	if err != nil {
		h.log.Error("Handler: Failed to get team", "error", err, "team_name", teamName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if team == nil {
		h.log.Info("Handler: Team not found", "team_name", teamName)
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	h.log.Info("Handler: Team retrieved successfully", "team_name", team.Name)
	c.JSON(http.StatusOK, team)
}
