package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailTaken = errors.New("email already registered")

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Repository interface {
	CreateUser(context.Context, string, string, string) (User, error)
	UserByEmail(context.Context, string) (User, string, error)
	CreateSession(context.Context, string, string, time.Time) error
	UserBySession(context.Context, string) (User, error)
	DeleteSession(context.Context, string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Register(ctx context.Context, name, email, password string) (User, error) {
	name, email = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(email))
	if name == "" || !strings.Contains(email, "@") || len(password) < 8 {
		return User{}, ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	return s.repo.CreateUser(ctx, name, email, string(hash))
}

func (s *Service) Login(ctx context.Context, email, password string) (User, error) {
	user, hash, err := s.repo.UserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) CreateSession(ctx context.Context, userID, tokenHash string, expires time.Time) error {
	return s.repo.CreateSession(ctx, userID, tokenHash, expires)
}
func (s *Service) CurrentUser(ctx context.Context, tokenHash string) (User, error) {
	return s.repo.UserBySession(ctx, tokenHash)
}
func (s *Service) Logout(ctx context.Context, tokenHash string) error {
	return s.repo.DeleteSession(ctx, tokenHash)
}
