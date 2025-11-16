package user

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/storage"
	"context"
)

type Service struct {
	storage storage.User
	log     logger.Logger
}

func NewUserService(r storage.User, log logger.Logger) Service {
	return Service{storage: r, log: log}
}

func (s *Service) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	s.log.Info("Creating user in service", "user_id", user.Id, "username", user.Username)

	createdUser, err := s.storage.CreateUser(ctx, user)
	if err != nil {
		s.log.Error("Failed to create user in service", "error", err, "user_id", user.Id)
		return nil, err
	}

	s.log.Info("Successfully created user in service", "user_id", createdUser.Id)
	return createdUser, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	s.log.Debug("Getting user by ID in service", "user_id", id)

	user, err := s.storage.GetUserByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get user by ID in service", "error", err, "user_id", id)
		return nil, err
	}

	s.log.Debug("Successfully retrieved user in service", "user_id", id, "found", user != nil)
	return user, nil
}
