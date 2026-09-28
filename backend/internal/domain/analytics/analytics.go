package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/trade-diary/backend/internal/domain/quality"
)

var ErrNotFound = errors.New("journal not found")

const (
	ReasonRoleMissing     = "role_missing"
	ReasonNoValues        = "no_values"
	ReasonNoWinLossValues = "no_win_loss_values"
	ReasonNoNegativePnL   = "no_negative_pnl"
)

type ColumnData struct {
	ID, Name, Type string
	Role           *string
	Options        []string
}
type RowData struct {
	ID     string
	Values map[string]json.RawMessage
}
type Dataset struct {
	ObservationCount int
	DefaultRisk      float64
	InitialDeposit   float64
	DefaultRR        float64
	Columns          []ColumnData
	Rows             []RowData
}
type Repository interface {
	Load(context.Context, string, string) (Dataset, error)
}

type DistributionValue struct {
	Value      string  `json:"value"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}
type SelectDistribution struct {
	ColumnID   string              `json:"columnId"`
	ColumnName string              `json:"columnName"`
	Total      int                 `json:"total"`
	Values     []DistributionValue `json:"values"`
}
type Metric struct {
	Available  bool     `json:"available"`
	Value      *float64 `json:"value,omitempty"`
	SampleSize int      `json:"sampleSize"`
	Reason     string   `json:"reason,omitempty"`
}
type Metrics struct {
	WinRate      Metric `json:"winRate"`
	TotalPnL     Metric `json:"totalPnl"`
	AveragePnL   Metric `json:"averagePnl"`
	ProfitFactor Metric `json:"profitFactor"`
	TotalR       Metric `json:"totalR"`
	AverageR     Metric `json:"averageR"`
}
type RowIssue struct {
	RowID string   `json:"rowId"`
	Codes []string `json:"codes"`
}
type DataQuality struct {
	CleanRows   int        `json:"cleanRows"`
	WarningRows int        `json:"warningRows"`
	Issues      []RowIssue `json:"issues"`
}
type TradeCounts struct {
	Wins      int `json:"wins"`
	Losses    int `json:"losses"`
	Breakeven int `json:"breakeven"`
}
type Report struct {
	ObservationCount    int                  `json:"observationCount"`
	SelectDistributions []SelectDistribution `json:"selectDistributions"`
	Metrics             Metrics              `json:"metrics"`
	DataQuality         DataQuality          `json:"dataQuality"`
	Trades              TradeCounts          `json:"trades"`
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Get(ctx context.Context, userID, journalID string) (Report, error) {
	dataset, err := s.repo.Load(ctx, userID, journalID)
	if err != nil {
		return Report{}, err
	}
	return Calculate(dataset), nil
}

func Calculate(dataset Dataset) Report {
	report := Report{ObservationCount: dataset.ObservationCount, SelectDistributions: []SelectDistribution{}, DataQuality: DataQuality{Issues: []RowIssue{}}, Metrics: Metrics{
		WinRate: unavailable(ReasonRoleMissing), TotalPnL: unavailable(ReasonRoleMissing), AveragePnL: unavailable(ReasonRoleMissing), ProfitFactor: unavailable(ReasonRoleMissing), TotalR: unavailable(ReasonRoleMissing), AverageR: unavailable(ReasonRoleMissing),
	}}
	roles := map[string]string{}
	for _, column := range dataset.Columns {
		if column.Type == "select" {
			report.SelectDistributions = append(report.SelectDistributions, distribution(column, dataset.Rows))
		}
		if column.Role != nil {
			if _, ok := roles[*column.Role]; !ok {
				roles[*column.Role] = column.ID
			}
		}
	}
	results := []string{}
	pnls := []float64{}
	rs := []float64{}
	for _, row := range dataset.Rows {
		riskPercent := numberValue(row.Values[roles["risk"]])
		if riskPercent == nil && dataset.DefaultRisk > 0 {
			riskPercent = &dataset.DefaultRisk
		}
		rr := numberValue(row.Values[roles["r"]])
		if rr == nil && dataset.DefaultRR > 0 {
			rr = &dataset.DefaultRR
		}
		values := quality.Values{Result: stringValue(row.Values[roles["trade_result"]]), PnL: numberValue(row.Values[roles["pnl"]]), R: rr, Risk: riskPercent, Deposit: dataset.InitialDeposit}
		issues := quality.Evaluate(values)
		if len(issues) > 0 {
			report.DataQuality.WarningRows++
			report.DataQuality.Issues = append(report.DataQuality.Issues, RowIssue{RowID: row.ID, Codes: issues})
		} else {
			report.DataQuality.CleanRows++
		}
		if values.Result != nil {
			results = append(results, *values.Result)
		} else if values.PnL != nil {
			results = appendOutcome(results, *values.PnL)
		}
		riskAmount := 0.0
		if values.Risk != nil && *values.Risk > 0 {
			riskAmount = dataset.InitialDeposit * *values.Risk / 100
		}
		if values.PnL != nil {
			pnls = append(pnls, *values.PnL)
		} else if values.Result != nil && values.R != nil && riskAmount > 0 {
			if pnl := pnlForOutcome(*values.Result, riskAmount, *values.R); pnl != nil {
				pnls = append(pnls, *pnl)
			}
		}
		if values.Result != nil && values.R != nil {
			if actualR := rForOutcome(*values.Result, *values.R); actualR != nil {
				rs = append(rs, *actualR)
			}
		} else if values.PnL != nil && riskAmount > 0 {
			rs = append(rs, *values.PnL/riskAmount)
		}
	}
	report.Metrics.WinRate = winRate(results)
	report.Trades = countTrades(results)
	report.Metrics.TotalPnL, report.Metrics.AveragePnL, report.Metrics.ProfitFactor = pnlMetrics(pnls)
	report.Metrics.TotalR, report.Metrics.AverageR = sumAndAverage(rs)
	return report
}
func pnlForOutcome(result string, riskAmount, rr float64) *float64 {
	actualR := rForOutcome(result, rr)
	if actualR == nil {
		return nil
	}
	value := riskAmount * *actualR
	return &value
}
func rForOutcome(result string, rr float64) *float64 {
	var value float64
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "win":
		value = rr
	case "loss":
		value = -1
	case "breakeven":
		value = 0
	default:
		return nil
	}
	return &value
}

func distribution(column ColumnData, rows []RowData) SelectDistribution {
	counts := map[string]int{}
	for _, option := range column.Options {
		counts[option] = 0
	}
	total := 0
	for _, row := range rows {
		if value := stringValue(row.Values[column.ID]); value != nil {
			counts[*value]++
			total++
		}
	}
	values := []DistributionValue{}
	seen := map[string]bool{}
	appendValue := func(value string) {
		percentage := 0.0
		if total > 0 {
			percentage = float64(counts[value]) * 100 / float64(total)
		}
		values = append(values, DistributionValue{Value: value, Count: counts[value], Percentage: percentage})
		seen[value] = true
	}
	for _, option := range column.Options {
		appendValue(option)
	}
	extra := []string{}
	for value := range counts {
		if !seen[value] {
			extra = append(extra, value)
		}
	}
	sort.Strings(extra)
	for _, value := range extra {
		appendValue(value)
	}
	return SelectDistribution{ColumnID: column.ID, ColumnName: column.Name, Total: total, Values: values}
}
func winRate(values []string) Metric {
	wins, losses := 0, 0
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "win":
			wins++
		case "loss":
			losses++
		}
	}
	total := wins + losses
	if total == 0 {
		return unavailable(ReasonNoWinLossValues)
	}
	return available(float64(wins)*100/float64(total), total)
}
func appendOutcome(values []string, result float64) []string {
	if result > 0 {
		return append(values, "win")
	}
	if result < 0 {
		return append(values, "loss")
	}
	return append(values, "breakeven")
}
func countTrades(values []string) TradeCounts {
	var counts TradeCounts
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "win":
			counts.Wins++
		case "loss":
			counts.Losses++
		case "breakeven":
			counts.Breakeven++
		}
	}
	return counts
}
func pnlMetrics(numbers []float64) (Metric, Metric, Metric) {
	if len(numbers) == 0 {
		m := unavailable(ReasonNoValues)
		return m, m, m
	}
	total, profit, loss := 0.0, 0.0, 0.0
	for _, v := range numbers {
		total += v
		if v > 0 {
			profit += v
		} else if v < 0 {
			loss -= v
		}
	}
	tm := available(total, len(numbers))
	am := available(total/float64(len(numbers)), len(numbers))
	if loss == 0 {
		return tm, am, Metric{Available: false, SampleSize: len(numbers), Reason: ReasonNoNegativePnL}
	}
	return tm, am, available(profit/loss, len(numbers))
}
func sumAndAverage(numbers []float64) (Metric, Metric) {
	if len(numbers) == 0 {
		m := unavailable(ReasonNoValues)
		return m, m
	}
	total := 0.0
	for _, v := range numbers {
		total += v
	}
	return available(total, len(numbers)), available(total/float64(len(numbers)), len(numbers))
}
func stringValue(raw json.RawMessage) *string {
	var v string
	if len(raw) == 0 || json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return &v
}
func numberValue(raw json.RawMessage) *float64 {
	var v float64
	if len(raw) == 0 || json.Unmarshal(raw, &v) != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}
func available(value float64, n int) Metric {
	return Metric{Available: true, Value: &value, SampleSize: n}
}
func unavailable(reason string) Metric { return Metric{Reason: reason} }
