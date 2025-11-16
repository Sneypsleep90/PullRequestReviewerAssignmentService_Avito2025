package models

type User struct {
	Id       string `db:"id" json:"id" binding:"required"`
	Username string `db:"username" json:"username" binding:"required"`
	IsActive bool   `db:"is_active" json:"is_active"`
	TeamName string `db:"team_name" json:"team_name"`
}
