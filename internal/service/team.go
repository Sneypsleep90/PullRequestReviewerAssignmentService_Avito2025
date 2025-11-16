package service

import (
	"avito-autumn-2025/internal/models"
	"context"
)

type Team interface {
	CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error)
	GetTeamWithMembers(ctx context.Context, teamName string) (*models.Team, error)
}
