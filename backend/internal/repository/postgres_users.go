package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"bus-tracker/internal/domain"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore { return &UserStore{pool: pool} }

// FindByEmail returns (nil, nil) when no such user exists.
func (s *UserStore) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	var role string
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, role::text, COALESCE(school_id, 0), COALESCE(password_hash, '')
		FROM users WHERE lower(email) = lower($1)`, email).
		Scan(&u.ID, &u.Name, &u.Email, &role, &u.SchoolID, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Role = domain.Role(role)
	return &u, nil
}

// DriverBusID returns the bus assigned to this driver, or 0 if none.
func (s *UserStore) DriverBusID(ctx context.Context, userID int64) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM buses WHERE driver_id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// CanWatchBus: drivers -> their own bus; parents -> subscribed buses;
// principals -> any bus on a route of their school.
func (s *UserStore) CanWatchBus(ctx context.Context, c domain.Claims, busID int64) (bool, error) {
	var q string
	var args []any
	switch c.Role {
	case domain.RoleDriver:
		q = `SELECT EXISTS (SELECT 1 FROM buses WHERE id = $2 AND driver_id = $1)`
		args = []any{c.UserID, busID}
	case domain.RoleParent:
		q = `SELECT EXISTS (SELECT 1 FROM parent_bus_subscriptions WHERE parent_id = $1 AND bus_id = $2)`
		args = []any{c.UserID, busID}
	case domain.RolePrincipal:
		q = `SELECT EXISTS (SELECT 1 FROM buses b JOIN routes r ON r.id = b.route_id
		     WHERE b.id = $2 AND r.school_id = $1)`
		args = []any{c.SchoolID, busID}
	default:
		return false, nil
	}
	var ok bool
	err := s.pool.QueryRow(ctx, q, args...).Scan(&ok)
	return ok, err
}
