package postgres

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamStorage struct {
	db          *pgxpool.Pool
	log         logger.Logger
	userStorage UserStorage
}

func NewTeamStorage(db *pgxpool.Pool, log logger.Logger) TeamStorage {
	return TeamStorage{
		db:          db,
		log:         log,
		userStorage: NewUserStorage(db, log),
	}
}

func (t *TeamStorage) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	t.log.Info("Creating team", "team_name", team.Name, "members_count", len(team.Users))

	tx, err := t.db.Begin(ctx)
	if err != nil {
		t.log.Error("Failed to begin transaction for team creation", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO teams (name) VALUES($1) RETURNING name`
	var createdTeam models.Team
	err = tx.QueryRow(ctx, query, team.Name).Scan(&createdTeam.Name)
	if err != nil {
		t.log.Error("Failed to create team", "error", err, "team_name", team.Name)
		return nil, fmt.Errorf("failed to create team: %w", err)
	}
	createdTeam.Id = createdTeam.Name

	for _, member := range team.Users {
		user := &models.User{
			Id:       member.Id,
			Username: member.Username,
			IsActive: member.IsActive,
		}

		upsertQuery := `
			INSERT INTO users (id, username, is_active, team_name) 
			VALUES($1, $2, $3, $4) 
			ON CONFLICT (id) DO UPDATE SET 
				username = EXCLUDED.username,
				is_active = EXCLUDED.is_active,
				team_name = EXCLUDED.team_name
		`
		_, err = tx.Exec(ctx, upsertQuery, user.Id, user.Username, user.IsActive, createdTeam.Name)
		if err != nil {
			t.log.Error("Failed to upsert team member", "error", err, "user_id", user.Id, "team_name", createdTeam.Name)
			return nil, fmt.Errorf("failed to upsert team member %s: %w", user.Id, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		t.log.Error("Failed to commit team creation transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	members, err := t.userStorage.GetUsersByTeam(ctx, createdTeam.Name)
	if err != nil {
		t.log.Error("Failed to load team members after creation", "error", err, "team_name", createdTeam.Name)
		return nil, fmt.Errorf("failed to load team members: %w", err)
	}
	createdTeam.Users = members

	t.log.Info("Successfully created team", "team_name", createdTeam.Name, "members_count", len(members))
	return &createdTeam, nil
}

func (t *TeamStorage) GetTeamWithMembers(ctx context.Context, teamName string) (*models.Team, error) {
	t.log.Debug("Getting team with members", "team_name", teamName)

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`
	err := t.db.QueryRow(ctx, checkQuery, teamName).Scan(&exists)
	if err != nil {
		t.log.Error("Failed to check team existence", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}

	if !exists {
		t.log.Warn("Team not found", "team_name", teamName)
		return nil, fmt.Errorf("team %s not found", teamName)
	}

	members, err := t.userStorage.GetUsersByTeam(ctx, teamName)
	if err != nil {
		t.log.Error("Failed to get team members", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	team := &models.Team{
		Id:    teamName,
		Name:  teamName,
		Users: members,
	}

	t.log.Debug("Successfully retrieved team with members", "team_name", teamName, "members_count", len(members))
	return team, nil
}
