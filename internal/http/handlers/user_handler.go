package handlers

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.User
	log         logger.Logger
}

func NewUserHandler(userService service.User, log logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	h.log.Debug("Handler: Creating user request")

	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Handler: Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.Id == "" || req.Username == "" {
		h.log.Error("Handler: Required fields missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and username are required"})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Handler: Failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Handler: User created successfully", "user_id", user.Id)
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	h.log.Debug("Handler: Getting user by ID request")

	userID := c.Param("id")
	if userID == "" {
		h.log.Error("Handler: user ID parameter is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID parameter is required"})
		return
	}

	h.log.Debug("Handler: Getting user", "user_id", userID)

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("Handler: Failed to get user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		h.log.Info("Handler: User not found", "user_id", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.log.Info("Handler: User retrieved successfully", "user_id", user.Id)
	c.JSON(http.StatusOK, user)
}
