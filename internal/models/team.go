package models

type Team struct {
	Id    string  `db:"id" json:"id"`
	Name  string  `db:"name" json:"team_name" binding:"required"`
	Users []*User `json:"members" binding:"dive"`
}
