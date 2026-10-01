package config

import "testing"

func TestHTTPAddressUsesDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")

	if got := FromEnv().HTTPAddress; got != ":8080" {
		t.Fatalf("HTTPAddress = %q, want %q", got, ":8080")
	}
}

func TestHTTPAddressUsesPlatformPort(t *testing.T) {
	t.Setenv("PORT", "3210")

	if got := FromEnv().HTTPAddress; got != ":3210" {
		t.Fatalf("HTTPAddress = %q, want %q", got, ":3210")
	}
}
