package entity

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
