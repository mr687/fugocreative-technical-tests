package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mr687/fugocreative-technical-tests/internal/model"
)

func TestOpenAIProviderUsesStructuredIndonesianOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected provider request path=%s auth=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		messages := request["messages"].([]any)
		systemMessage := messages[0].(map[string]any)
		if !strings.Contains(systemMessage["content"].(string), "bahasa Indonesia") {
			t.Error("LLM system prompt must require Indonesian responses")
		}
		format := request["response_format"].(map[string]any)
		if format["type"] != "json_schema" {
			t.Errorf("response format type = %v, want json_schema", format["type"])
		}
		schema, ok := format["json_schema"].(map[string]any)
		if !ok {
			t.Errorf("response format is missing JSON Schema: %+v", format)
		} else {
			if schema["strict"] != true || schema["name"] != "brief_analysis" {
				t.Errorf("unexpected JSON Schema metadata: %+v", schema)
			}
			root, ok := schema["schema"].(map[string]any)
			if !ok || root["type"] != "object" || root["additionalProperties"] != false {
				t.Errorf("unexpected JSON Schema object definition: %+v", schema["schema"])
			} else {
				properties, ok := root["properties"].(map[string]any)
				if !ok || len(properties) != 5 {
					t.Errorf("schema properties = %+v, want all five AI result fields", root["properties"])
				} else {
					category, ok := properties["service_category"].(map[string]any)
					enum, enumOK := category["enum"].([]any)
					if !ok || !enumOK || len(enum) != 6 {
						t.Errorf("service_category schema missing allowed enum values: %+v", properties["service_category"])
					}
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"summary\":\"Situs perusahaan\",\"service_category\":\"Website\",\"priority\":\"Medium\",\"priority_reason\":\"Lingkup jelas\",\"missing_info\":[]}"}}]}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL+"/v1", "model-x", "test-key")
	result, err := provider.Process(context.Background(), model.Brief{CompanyName: "Studio"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ServiceCategory != "Website" || result.Priority != "Medium" || result.Summary != "Situs perusahaan" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestOpenAIProviderRejectsInvalidJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"summary\":\"ok\",\"service_category\":\"unknown\",\"priority\":\"Low\",\"priority_reason\":\"why\",\"missing_info\":[]}"}}]}`))
	}))
	defer server.Close()
	_, err := NewOpenAIProvider(server.URL, "model-x", "test-key").Process(context.Background(), model.Brief{})
	if err == nil || !strings.Contains(err.Error(), "service_category") {
		t.Fatalf("error = %v, want invalid schema error", err)
	}
}

func TestOpenAIProviderGeneratesStructuredEmailDraft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		format, ok := request["response_format"].(map[string]any)
		if !ok || format["type"] != "json_schema" {
			t.Fatalf("email draft request must use JSON schema, got %v", request["response_format"])
		}
		jsonSchema := format["json_schema"].(map[string]any)
		schema := jsonSchema["schema"].(map[string]any)
		properties := schema["properties"].(map[string]any)
		if jsonSchema["name"] != "email_draft" || schema["additionalProperties"] != false ||
			len(properties) != 2 || properties["subject"] == nil || properties["body"] == nil {
			t.Errorf("unexpected email draft schema: %+v", jsonSchema)
		}
		messages := request["messages"].([]any)
		systemMessage := messages[0].(map[string]any)
		systemPrompt := systemMessage["content"].(string)
		for _, required := range []string{"Fugo Creative Group", "Bahasa Indonesia", "discovery call", "jangan gunakan markdown", "data tak tepercaya"} {
			if !strings.Contains(strings.ToLower(systemPrompt), strings.ToLower(required)) {
				t.Errorf("system prompt missing %q", required)
			}
		}
		humanMessage := messages[1].(map[string]any)
		if !strings.Contains(humanMessage["content"].(string), `"summary":"Ringkasan"`) {
			t.Error("email draft request must include AI summary")
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"subject\":\"Tindak lanjut proyek website\",\"body\":\"Yth. Ayu, terima kasih atas brief Anda.\"}"}}]}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL+"/v1", "model-x", "test-key")
	draft, err := provider.GenerateEmailDraft(context.Background(), model.EmailDraftContext{
		ContactName: "Ayu", CompanyName: "Studio", NeedDescription: "Website",
		Summary: "Ringkasan", MissingInfo: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Subject != "Tindak lanjut proyek website" || draft.Body != "Yth. Ayu, terima kasih atas brief Anda." {
		t.Fatalf("draft = %+v", draft)
	}
}
