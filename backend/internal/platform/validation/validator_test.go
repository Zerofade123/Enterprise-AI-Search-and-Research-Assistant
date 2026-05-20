package validation

import "testing"

type sample struct {
	Name string `validate:"required"`
}

func TestValidateStruct(t *testing.T) {
	if err := ValidateStruct(sample{Name: ""}); err == nil {
		t.Fatalf("expected validation error")
	}

	if err := ValidateStruct(sample{Name: "ok"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
