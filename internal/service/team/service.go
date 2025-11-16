package team

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"avito-autumn-2025/internal/storage"
	"context"
)

type Service struct {
	storage storage.Team
	log     logger.Logger
}

func NewTeamService(r storage.Team, log logger.Logger) Service {
	return Service{storage: r, log: log}
}

func (s *Service) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	s.log.Info("Creating team in service", "team_name", team.Name, "members_count", len(team.Users))

	createdTeam, err := s.storage.CreateTeam(ctx, team)
	if err != nil {
		s.log.Error("Failed to create team in service", "error", err, "team_name", team.Name)
		return nil, err
	}

	s.log.Info("Successfully created team in service", "team_name", createdTeam.Name, "members_count", len(createdTeam.Users))
	return createdTeam, nil
}

func (s *Service) GetTeamWithMembers(ctx context.Context, teamName string) (*models.Team, error) {
	s.log.Debug("Getting team with members in service", "team_name", teamName)

	team, err := s.storage.GetTeamWithMembers(ctx, teamName)
	if err != nil {
		s.log.Error("Failed to get team with members in service", "error", err, "team_name", teamName)
		return nil, err
	}

	s.log.Debug("Successfully retrieved team with members in service", "team_name", teamName, "members_count", len(team.Users))
	return team, nil
}
