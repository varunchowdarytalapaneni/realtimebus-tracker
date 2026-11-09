package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"bus-tracker/internal/domain"
)

type UserRepo interface {
	FindByEmail(context.Context, string) (*domain.User, error)
	DriverBusID(context.Context, int64) (int64, error)
	CanWatchBus(context.Context, domain.Claims, int64) (bool, error)
}

type AuthService struct {
	users  UserRepo
	secret []byte
	ttl    time.Duration
}

func NewAuthService(users UserRepo, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, secret: []byte(secret), ttl: ttl}
}

type LoginResult struct {
	Token string
	User  domain.User
	BusID int64
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

type tokenClaims struct {
	Role     domain.Role `json:"role"`
	SchoolID int64       `json:"school_id"`
	BusID    int64       `json:"bus_id,omitempty"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	hash := dummyHash
	if u != nil && u.PasswordHash != "" {
		hash = []byte(u.PasswordHash)
	}
	if u == nil || u.PasswordHash == "" || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	var busID int64
	if u.Role == domain.RoleDriver {
		if busID, err = s.users.DriverBusID(ctx, u.ID); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		Role: u.Role, SchoolID: u.SchoolID, BusID: busID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.FormatInt(u.ID, 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	})
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}
	u.PasswordHash = ""
	return &LoginResult{Token: signed, User: *u, BusID: busID}, nil
}

func (s *AuthService) ParseToken(tokenStr string) (domain.Claims, error) {
	var tc tokenClaims
	tok, err := jwt.ParseWithClaims(tokenStr, &tc, func(*jwt.Token) (any, error) { return s.secret, nil }, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !tok.Valid {
		return domain.Claims{}, domain.ErrInvalidCredentials
	}
	uid, err := strconv.ParseInt(tc.Subject, 10, 64)
	if err != nil {
		return domain.Claims{}, domain.ErrInvalidCredentials
	}
	return domain.Claims{UserID: uid, Role: tc.Role, SchoolID: tc.SchoolID, BusID: tc.BusID}, nil
}

func (s *AuthService) CanWatch(ctx context.Context, c domain.Claims, busID int64) (bool, error) {
	return s.users.CanWatchBus(ctx, c, busID)
}
