package observation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"time"
)

var ErrNotFound = errors.New("row or column not found")
var ErrInvalidValue = errors.New("invalid cell value")
var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type Row struct {
	ID        string                     `json:"id"`
	JournalID string                     `json:"journalId"`
	Position  int                        `json:"position"`
	Values    map[string]json.RawMessage `json:"values"`
	CreatedAt time.Time                  `json:"createdAt"`
	UpdatedAt time.Time                  `json:"updatedAt"`
}
type Repository interface {
	List(context.Context, string, string) ([]Row, error)
	Create(context.Context, string, string) (Row, error)
	Delete(context.Context, string, string, string) error
	ColumnType(context.Context, string, string, string) (string, []string, error)
	SetCell(context.Context, string, string, string, string, json.RawMessage) error
	ClearCell(context.Context, string, string, string, string) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID, journalID string) ([]Row, error) {
	return s.repo.List(ctx, userID, journalID)
}
func (s *Service) Create(ctx context.Context, userID, journalID string) (Row, error) {
	return s.repo.Create(ctx, userID, journalID)
}
func (s *Service) Delete(ctx context.Context, userID, journalID, rowID string) error {
	return s.repo.Delete(ctx, userID, journalID, rowID)
}
func (s *Service) SetCell(ctx context.Context, userID, journalID, rowID, columnID string, value json.RawMessage) error {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return s.repo.ClearCell(ctx, userID, journalID, rowID, columnID)
	}
	typ, options, err := s.repo.ColumnType(ctx, userID, journalID, columnID)
	if err != nil {
		return err
	}
	if !validValue(typ, options, value) {
		return ErrInvalidValue
	}
	return s.repo.SetCell(ctx, userID, journalID, rowID, columnID, value)
}
func (s *Service) ClearCell(ctx context.Context, userID, journalID, rowID, columnID string) error {
	return s.repo.ClearCell(ctx, userID, journalID, rowID, columnID)
}
func validValue(typ string, options []string, raw json.RawMessage) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return false
	}
	switch typ {
	case "text":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "date":
		v, ok := value.(string)
		if !ok || !datePattern.MatchString(v) {
			return false
		}
		_, err := time.Parse("2006-01-02", v)
		return err == nil
	case "select":
		v, ok := value.(string)
		if !ok {
			return false
		}
		for _, option := range options {
			if v == option {
				return true
			}
		}
		return false
	default:
		return false
	}
}
