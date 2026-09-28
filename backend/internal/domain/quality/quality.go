package quality

import (
	"math"
	"strings"
)

const (
	ResultPnLConflict = "result_pnl_conflict"
	ResultRConflict   = "result_r_conflict"
	PnLRiskRConflict  = "pnl_risk_r_conflict"
	RiskNotPositive   = "risk_not_positive"
	RRNotPositive     = "rr_not_positive"
)

type Values struct {
	Result  *string
	PnL     *float64
	R       *float64
	Risk    *float64
	Deposit float64
}

func Evaluate(values Values) []string {
	issues := make([]string, 0)
	if values.Risk != nil && (*values.Risk <= 0 || *values.Risk > 100) {
		issues = append(issues, RiskNotPositive)
	}
	if values.R != nil && *values.R <= 0 {
		issues = append(issues, RRNotPositive)
	}
	if values.Result != nil && values.PnL != nil && signConflict(*values.Result, *values.PnL) {
		issues = append(issues, ResultPnLConflict)
	}
	if values.PnL != nil && values.R != nil && values.Risk != nil && values.Result != nil && *values.Risk > 0 && values.Deposit > 0 {
		riskAmount := values.Deposit * *values.Risk / 100
		expected := riskAmount * *values.R
		switch normalizeResult(*values.Result) {
		case "loss":
			expected = -riskAmount
		case "breakeven":
			expected = 0
		}
		tolerance := math.Max(0.01, math.Abs(expected)*0.001)
		if math.Abs(*values.PnL-expected) > tolerance {
			issues = append(issues, PnLRiskRConflict)
		}
	}
	return issues
}

func signConflict(result string, value float64) bool {
	switch normalizeResult(result) {
	case "win":
		return value < 0
	case "loss":
		return value > 0
	default:
		return false
	}
}

func normalizeResult(result string) string {
	result = strings.TrimSpace(result)
	if strings.EqualFold(result, "win") {
		return "win"
	}
	if strings.EqualFold(result, "loss") {
		return "loss"
	}
	if strings.EqualFold(result, "breakeven") {
		return "breakeven"
	}
	return ""
}
