package request

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type loginRequest struct {
	Username string `json:"username" validate:"required,notblank"`
}

func TestDecodeAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantFields int
		wantErr    bool
	}{
		{"valid JSON", `{"username":"admin"}`, 0, false},
		{"required field", `{"username":""}`, 1, false},
		{"blank field", `{"username":"  "}`, 1, false},
		{"unknown field", `{"username":"admin","extra":true}`, 0, true},
		{"trailing JSON", `{"username":"admin"}{}`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			fields, err := DecodeAndValidate(recorder, req, &loginRequest{})
			if (err != nil) != tt.wantErr || len(fields) != tt.wantFields {
				t.Fatalf("fields=%v err=%v, want fields=%d err=%t", fields, err, tt.wantFields, tt.wantErr)
			}
		})
	}
}
