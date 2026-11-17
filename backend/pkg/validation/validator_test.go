package validation

import (
	"strings"
	"testing"
)

func TestStructReturnsIndonesianErrors(t *testing.T) {
	type request struct {
		Username string `json:"username" validate:"required"`
	}
	got := Struct(request{})
	if len(got) != 1 {
		t.Fatalf("got %d errors, want 1", len(got))
	}
	if got[0].Field != "username" || !strings.Contains(strings.ToLower(got[0].Message), "wajib") {
		t.Fatalf("got %+v, want Indonesian required-field error for username", got[0])
	}
}
