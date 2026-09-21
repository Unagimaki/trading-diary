package column

import (
	"errors"
	"testing"
)

func TestValidateColumn(t *testing.T) {
	role := "pnl"
	v, err := validate(Values{Name: "  PnL ", Type: "number", Role: &role, Options: []string{"ignored"}})
	if err != nil {
		t.Fatal(err)
	}
	if v.Name != "PnL" || len(v.Options) != 0 {
		t.Fatalf("unexpected value: %#v", v)
	}
}
func TestValidateRejectsIncompatibleRole(t *testing.T) {
	role := "pnl"
	_, err := validate(Values{Name: "Result", Type: "select", Role: &role})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
