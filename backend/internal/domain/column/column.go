package column

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("column not found")
var ErrInvalid = errors.New("invalid column")
var types = map[string]bool{"text": true, "number": true, "select": true, "date": true, "boolean": true, "image": true}
var roles = map[string]string{"trade_result": "select", "pnl": "number", "r": "number"}

type Column struct {
	ID        string    `json:"id"`
	JournalID string    `json:"journalId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Role      *string   `json:"role"`
	Options   []string  `json:"options"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Values struct {
	Name     string
	Type     string
	Role     *string
	Options  []string
	Position *int
}
type Repository interface {
	List(context.Context, string, string) ([]Column, error)
	Get(context.Context, string, string, string) (Column, error)
	Create(context.Context, string, string, Values) (Column, error)
	Update(context.Context, string, string, string, Values) (Column, error)
	Delete(context.Context, string, string, string) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID, journalID string) ([]Column, error) {
	return s.repo.List(ctx, userID, journalID)
}
func (s *Service) Create(ctx context.Context, userID, journalID string, v Values) (Column, error) {
	v, err := validate(v)
	if err != nil {
		return Column{}, err
	}
	return s.repo.Create(ctx, userID, journalID, v)
}
func (s *Service) Update(ctx context.Context, userID, journalID, id string, v Values) (Column, error) {
	v, err := validate(v)
	if err != nil {
		return Column{}, err
	}
	return s.repo.Update(ctx, userID, journalID, id, v)
}
func (s *Service) Delete(ctx context.Context, userID, journalID, id string) error {
	return s.repo.Delete(ctx, userID, journalID, id)
}
func validate(v Values) (Values, error) {
	v.Name = strings.TrimSpace(v.Name)
	if len([]rune(v.Name)) < 1 || len([]rune(v.Name)) > 120 || !types[v.Type] {
		return v, ErrInvalid
	}
	if v.Role != nil {
		role := strings.TrimSpace(*v.Role)
		if role == "" {
			v.Role = nil
		} else if expected, ok := roles[role]; !ok || expected != v.Type {
			return v, ErrInvalid
		} else {
			v.Role = &role
		}
	}
	if v.Type != "select" {
		v.Options = []string{}
	} else {
		seen := map[string]bool{}
		clean := make([]string, 0, len(v.Options))
		for _, option := range v.Options {
			option = strings.TrimSpace(option)
			if option != "" && !seen[option] {
				seen[option] = true
				clean = append(clean, option)
			}
		}
		v.Options = clean
	}
	return v, nil
}
