package observation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/trade-diary/backend/internal/domain/quality"
)

var ErrNotFound = errors.New("row or column not found")
var ErrInvalidValue = errors.New("invalid cell value")
var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

type Row struct {
	ID        string                     `json:"id"`
	JournalID string                     `json:"journalId"`
	Position  int                        `json:"position"`
	Values    map[string]json.RawMessage `json:"values"`
	Warnings  []string                   `json:"warnings"`
	CreatedAt time.Time                  `json:"createdAt"`
	UpdatedAt time.Time                  `json:"updatedAt"`
}
type TradingCell struct {
	ColumnID string
	Value    json.RawMessage
	Source   string
	Options  []string
}
type TradingDefaults struct {
	Deposit     float64
	RiskPercent float64
	RR          float64
}
type Repository interface {
	List(context.Context, string, string) ([]Row, error)
	RoleColumns(context.Context, string, string) (map[string]string, error)
	TradingDefaults(context.Context, string, string) (TradingDefaults, error)
	Create(context.Context, string, string, string) (Row, error)
	Delete(context.Context, string, string, string) error
	ColumnType(context.Context, string, string, string) (string, string, []string, error)
	SetCell(context.Context, string, string, string, string, json.RawMessage) error
	ClearCell(context.Context, string, string, string, string) error
	TradingContext(context.Context, string, string, string) (map[string]TradingCell, TradingDefaults, error)
	ReplaceCalculated(context.Context, string, string, string, map[string]json.RawMessage) error
}
type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, userID, journalID string) ([]Row, error) {
	rows, err := s.repo.List(ctx, userID, journalID)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.RoleColumns(ctx, userID, journalID)
	if err != nil {
		return nil, err
	}
	defaults, err := s.repo.TradingDefaults(ctx, userID, journalID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		riskPercent := numberValue(rows[i].Values[roles["risk"]])
		if riskPercent == nil && defaults.RiskPercent > 0 {
			riskPercent = &defaults.RiskPercent
		}
		rows[i].Warnings = quality.Evaluate(quality.Values{
			Result:  stringValue(rows[i].Values[roles["trade_result"]]),
			PnL:     numberValue(rows[i].Values[roles["pnl"]]),
			R:       numberValue(rows[i].Values[roles["r"]]),
			Risk:    riskPercent,
			Deposit: defaults.Deposit,
		})
	}
	return rows, nil
}
func (s *Service) Create(ctx context.Context, userID, journalID, localDate string) (Row, error) {
	if _, err := time.Parse("2006-01-02", localDate); err != nil {
		localDate = time.Now().Format("2006-01-02")
	}
	row, err := s.repo.Create(ctx, userID, journalID, localDate)
	if err != nil {
		return row, err
	}
	return row, s.recalculate(ctx, userID, journalID, row.ID)
}
func (s *Service) Delete(ctx context.Context, userID, journalID, rowID string) error {
	return s.repo.Delete(ctx, userID, journalID, rowID)
}
func (s *Service) RecalculateJournal(ctx context.Context, userID, journalID string) error {
	rows, err := s.repo.List(ctx, userID, journalID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := s.recalculate(ctx, userID, journalID, row.ID); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) SetCell(ctx context.Context, userID, journalID, rowID, columnID string, value json.RawMessage) error {
	if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return s.ClearCell(ctx, userID, journalID, rowID, columnID)
	}
	typ, role, options, err := s.repo.ColumnType(ctx, userID, journalID, columnID)
	if err != nil {
		return err
	}
	if !validValue(typ, options, value) || !validRoleValue(role, value) {
		return ErrInvalidValue
	}
	if err := s.repo.SetCell(ctx, userID, journalID, rowID, columnID, value); err != nil {
		return err
	}
	return s.recalculate(ctx, userID, journalID, rowID)
}

func validRoleValue(role string, raw json.RawMessage) bool {
	if role != "risk" && role != "r" {
		return true
	}
	value := numberValue(raw)
	if value == nil || *value <= 0 {
		return false
	}
	return role != "risk" || *value <= 100
}
func (s *Service) ClearCell(ctx context.Context, userID, journalID, rowID, columnID string) error {
	if err := s.repo.ClearCell(ctx, userID, journalID, rowID, columnID); err != nil {
		return err
	}
	return s.recalculate(ctx, userID, journalID, rowID)
}
func (s *Service) recalculate(ctx context.Context, userID, journalID, rowID string) error {
	cells, defaults, err := s.repo.TradingContext(ctx, userID, journalID, rowID)
	if err != nil {
		return err
	}
	riskPercent := defaults.RiskPercent
	if cell, ok := cells["risk"]; ok {
		if value := numberValue(cell.Value); value != nil {
			riskPercent = *value
		}
	}
	derived := make(map[string]json.RawMessage)
	if riskCell, ok := cells["risk"]; ok && riskCell.Source != "manual" {
		derived[riskCell.ColumnID] = numberJSON(defaults.RiskPercent)
	}
	rr := defaults.RR
	rCell, hasR := cells["r"]
	if hasR {
		if value := numberValue(rCell.Value); value != nil {
			rr = *value
		}
		if rCell.Source != "manual" {
			derived[rCell.ColumnID] = numberJSON(defaults.RR)
			rr = defaults.RR
		}
	}
	pnlCell, hasPnL := cells["pnl"]
	resultCell, hasResult := cells["trade_result"]
	if hasResult {
		if result := stringValue(resultCell.Value); result != nil && hasPnL && defaults.Deposit > 0 && riskPercent > 0 && rr > 0 {
			if pnl := pnlForOutcome(*result, defaults.Deposit*riskPercent/100, rr); pnl != nil {
				derived[pnlCell.ColumnID] = numberJSON(*pnl)
			}
		}
	}
	if (!hasResult || resultCell.Source != "manual") && hasPnL && pnlCell.Source == "manual" {
		if value := numberValue(pnlCell.Value); value != nil {
			if hasResult {
				if result := outcomeOption(resultCell.Options, *value); result != nil {
					derived[resultCell.ColumnID] = stringJSON(*result)
				}
			}
		}
	}
	return s.repo.ReplaceCalculated(ctx, userID, journalID, rowID, derived)
}

func pnlForOutcome(result string, riskAmount, rr float64) *float64 {
	var value float64
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "win":
		value = riskAmount * rr
	case "loss":
		value = -riskAmount
	case "breakeven":
		value = 0
	default:
		return nil
	}
	return &value
}

func outcomeOption(options []string, value float64) *string {
	wanted := "Breakeven"
	if value > 0 {
		wanted = "Win"
	} else if value < 0 {
		wanted = "Loss"
	}
	for _, option := range options {
		if strings.EqualFold(strings.TrimSpace(option), wanted) {
			result := option
			return &result
		}
	}
	return nil
}
func numberJSON(value float64) json.RawMessage { raw, _ := json.Marshal(value); return raw }
func stringJSON(value string) json.RawMessage  { raw, _ := json.Marshal(value); return raw }
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

func stringValue(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}

func numberValue(raw json.RawMessage) *float64 {
	if len(raw) == 0 {
		return nil
	}
	var value float64
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}
