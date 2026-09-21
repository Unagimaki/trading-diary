package journal

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("journal not found")
var ErrInvalidName = errors.New("invalid journal name")

type Journal struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository interface {
	List(context.Context, string) ([]Journal, error)
	Create(context.Context, string, string) (Journal, error)
	Get(context.Context, string, string) (Journal, error)
	Rename(context.Context, string, string, string) (Journal, error)
	Delete(context.Context, string, string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID string) ([]Journal, error) {
	return s.repo.List(ctx, userID)
}
func (s *Service) Get(ctx context.Context, userID, id string) (Journal, error) {
	return s.repo.Get(ctx, userID, id)
}
func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}
func (s *Service) Create(ctx context.Context, userID, name string) (Journal, error) {
	name, err := validName(name)
	if err != nil {
		return Journal{}, err
	}
	return s.repo.Create(ctx, userID, name)
}
func (s *Service) Rename(ctx context.Context, userID, id, name string) (Journal, error) {
	name, err := validName(name)
	if err != nil {
		return Journal{}, err
	}
	return s.repo.Rename(ctx, userID, id, name)
}
func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 1 || len([]rune(name)) > 120 {
		return "", ErrInvalidName
	}
	return name, nil
}
