package quality

import "testing"

func pointer[T any](value T) *T { return &value }

func TestEvaluateContradictions(t *testing.T) {
	issues := Evaluate(Values{
		Result:  pointer("Win"),
		PnL:     pointer(-100.0),
		R:       pointer(2.0),
		Risk:    pointer(1.0),
		Deposit: 10000,
	})
	if len(issues) != 2 || issues[0] != ResultPnLConflict || issues[1] != PnLRiskRConflict {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}

func TestEvaluateConsistentValues(t *testing.T) {
	issues := Evaluate(Values{Result: pointer("Loss"), PnL: pointer(-100.0), R: pointer(2.0), Risk: pointer(1.0), Deposit: 10000})
	if len(issues) != 0 {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}
