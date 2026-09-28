package analytics

import (
	"encoding/json"
	"math"
	"testing"
)

func role(value string) *string { return &value }
func row(id string, values map[string]string) RowData {
	raw := make(map[string]json.RawMessage, len(values))
	for key, value := range values {
		raw[key] = json.RawMessage(value)
	}
	return RowData{ID: id, Values: raw}
}

func TestCalculateReportAndDerivedR(t *testing.T) {
	report := Calculate(Dataset{ObservationCount: 3, InitialDeposit: 10000, DefaultRisk: 1, DefaultRR: 2, Columns: []ColumnData{
		{ID: "result", Name: "Result", Type: "select", Role: role("trade_result"), Options: []string{"Win", "Loss"}},
		{ID: "pnl", Name: "PnL", Type: "number", Role: role("pnl")}, {ID: "risk", Name: "Risk", Type: "number", Role: role("risk")}, {ID: "rr", Name: "RR", Type: "number", Role: role("r")},
	}, Rows: []RowData{
		row("1", map[string]string{"result": `"Win"`, "pnl": "200", "risk": "1", "rr": "2"}),
		row("2", map[string]string{"result": `"Loss"`, "pnl": "-100", "risk": "1", "rr": "3"}),
		row("3", map[string]string{"result": `"Win"`, "pnl": "300", "risk": "1", "rr": "3"}),
	}})
	assertMetric(t, report.Metrics.WinRate, 200.0/3, 3)
	assertMetric(t, report.Metrics.TotalPnL, 400, 3)
	assertMetric(t, report.Metrics.ProfitFactor, 5, 3)
	assertMetric(t, report.Metrics.TotalR, 4, 3)
	assertMetric(t, report.Metrics.AverageR, 4.0/3, 3)
	if report.DataQuality.WarningRows != 0 || report.SelectDistributions[0].Total != 3 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestOutcomeDeterminesSignedR(t *testing.T) {
	report := Calculate(Dataset{ObservationCount: 2, InitialDeposit: 1000, DefaultRisk: 10, DefaultRR: 2, Columns: []ColumnData{
		{ID: "result", Type: "select", Role: role("trade_result")}, {ID: "r", Type: "number", Role: role("r")},
	}, Rows: []RowData{row("win", map[string]string{"result": `"Win"`, "r": "2"}), row("loss", map[string]string{"result": `"Loss"`, "r": "3"})}})
	if report.DataQuality.WarningRows != 0 {
		t.Fatalf("unexpected quality: %#v", report.DataQuality)
	}
	assertMetric(t, report.Metrics.WinRate, 50, 2)
	assertMetric(t, report.Metrics.TotalR, 1, 2)
	assertMetric(t, report.Metrics.TotalPnL, 100, 2)
}

func assertMetric(t *testing.T, metric Metric, want float64, sampleSize int) {
	t.Helper()
	if !metric.Available || metric.Value == nil || math.Abs(*metric.Value-want) > 0.000001 || metric.SampleSize != sampleSize {
		t.Fatalf("metric %#v, want %v (%d)", metric, want, sampleSize)
	}
}
