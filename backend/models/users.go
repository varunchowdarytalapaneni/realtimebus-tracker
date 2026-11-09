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

// Claims is what we trust after a token is verified. BusID is set for drivers only.
type Claims struct {
	UserID   int64
	Role     Role
	SchoolID int64
	BusID    int64
}