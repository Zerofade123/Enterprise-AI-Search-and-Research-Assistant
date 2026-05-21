package service

import "testing"

func TestPasswordStrength(t *testing.T) {
	if err := ValidatePasswordStrength("weak"); err == nil {
		t.Fatalf("expected weak password error")
	}

	if err := ValidatePasswordStrength("StrongPassw0rd!"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
