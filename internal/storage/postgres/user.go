package postgres

import (
	"avito-autumn-2025/internal/logger"
	"avito-autumn-2025/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStorage struct {
	db  *pgxpool.Pool
	log logger.Logger
}

func NewUserStorage(db *pgxpool.Pool, log logger.Logger) UserStorage {
	return UserStorage{db: db, log: log}
}

func (u *UserStorage) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	u.log.Info("Creating user", "user_id", user.Id, "username", user.Username, "is_active", user.IsActive)

	query := `INSERT INTO users (id, username, is_active) VALUES($1, $2, $3)  RETURNING id, username, is_active`
	var createdUser models.User
	err := u.db.QueryRow(ctx, query, user.Id, user.Username, user.IsActive).Scan(&createdUser.Id, &createdUser.Username, &createdUser.IsActive)
	if err != nil {
		u.log.Error("Failed to create user", "error", err, "user_id", user.Id)
		return nil, err
	}

	u.log.Info("Successfully created user", "user_id", createdUser.Id)
	return &createdUser, nil
}

func (u *UserStorage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	u.log.Debug("Getting user by ID", "user_id", id)

	data := &models.User{}
	query := `SELECT id, username, is_active, team_name FROM users WHERE id = $1`
	row := u.db.QueryRow(ctx, query, id)
	var teamName sql.NullString
	err := row.Scan(&data.Id, &data.Username, &data.IsActive, &teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			u.log.Debug("User not found", "user_id", id)
			return nil, nil
		}
		u.log.Error("Failed to get user by ID", "error", err, "user_id", id)
		return nil, err
	}

	if teamName.Valid {
		data.TeamName = teamName.String
	}

	u.log.Debug("Successfully retrieved user", "user_id", id)
	return data, nil
}

func (u *UserStorage) GetUsersByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	u.log.Debug("Getting users by team", "team_name", teamName)

	query := `SELECT id, username, is_active, team_name FROM users WHERE team_name = $1 ORDER BY username`
	rows, err := u.db.Query(ctx, query, teamName)
	if err != nil {
		u.log.Error("Failed to get users by team", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to get users by team: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName); err != nil {
			u.log.Error("Failed to scan user", "error", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	u.log.Debug("Successfully retrieved users by team", "team_name", teamName, "count", len(users))
	return users, nil
}

func (u *UserStorage) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	u.log.Info("Setting user active status", "user_id", userID, "is_active", isActive)

	query := `UPDATE users SET is_active = $1 WHERE id = $2`
	result, err := u.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		u.log.Error("Failed to set user active status", "error", err, "user_id", userID)
		return fmt.Errorf("failed to set user active status: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		u.log.Warn("User not found for active status update", "user_id", userID)
		return fmt.Errorf("user %s not found", userID)
	}

	u.log.Info("Successfully updated user active status", "user_id", userID, "is_active", isActive)
	return nil
}
