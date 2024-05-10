package authRoutes_test

import "testing"

func TestRequestNewAccount(t *testing.T) {
	exp := 4

	got := 5

	if exp != got {
		t.Errorf("Expected %d, got %d", exp, got)
	}
}
