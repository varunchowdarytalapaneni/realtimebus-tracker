package domain

import "errors"

type Role string

const (
	RoleDriver    Role = "driver"
	RoleParent    Role = "parent"
	RolePrincipal Role = "principal"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         Role   `json:"role"`
	SchoolID     int64  `json:"school_id"`
	PasswordHash string `json:"-"`
}

type Claims struct {
	UserID   int64
	Role     Role
	SchoolID int64
	BusID    int64
}
