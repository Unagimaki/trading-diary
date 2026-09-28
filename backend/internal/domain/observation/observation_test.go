package observation

import (
	"encoding/json"
	"testing"
)

func TestValidValue(t *testing.T) {
	tests := []struct {
		typ     string
		options []string
		raw     string
		want    bool
	}{{"text", nil, `"note"`, true}, {"text", nil, `12`, false}, {"number", nil, `12.5`, true}, {"number", nil, `"12"`, false}, {"boolean", nil, `true`, true}, {"date", nil, `"2026-02-29"`, false}, {"date", nil, `"2026-02-28"`, true}, {"select", []string{"Win", "Loss"}, `"Win"`, true}, {"select", []string{"Win", "Loss"}, `"Other"`, false}, {"image", nil, `"file"`, false}}
	for _, tt := range tests {
		if got := validValue(tt.typ, tt.options, json.RawMessage(tt.raw)); got != tt.want {
			t.Errorf("%s %s: got %v", tt.typ, tt.raw, got)
		}
	}
}

func TestValidRoleValue(t *testing.T) {
	tests := []struct {
		role string
		raw  string
		want bool
	}{
		{"risk", `1`, true},
		{"risk", `0.01`, true},
		{"risk", `0`, false},
		{"risk", `-1`, false},
		{"r", `-1`, false},
		{"pnl", `-100`, true},
	}
	for _, tt := range tests {
		if got := validRoleValue(tt.role, json.RawMessage(tt.raw)); got != tt.want {
			t.Errorf("%s %s: got %v, want %v", tt.role, tt.raw, got, tt.want)
		}
	}
}
