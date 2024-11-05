package model

import "time"

type OrgDTO struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type Org struct {
	Name        string    `db:"name"`
	Owner       string    `db:"owner"`
	Members     []string  `db:"members"`     // Using []string to represent SET<TEXT>
	Permissions []string  `db:"permissions"` // Using []string for the permissions set
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
type Permission struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
