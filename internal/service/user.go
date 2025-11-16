package service

import (
	"avito-autumn-2025/internal/models"
	"context"
)

type User interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}
