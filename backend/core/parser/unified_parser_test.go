package parser

import (
	"strings"
	"testing"

	"RequestProbe/backend/models"
)

func TestUnifiedRequestParser_parseURLAndParams(t *testing.T) {
	parser := NewUnifiedRequestParser()

	baseURL, params := parser.parseURLAndParams("https://example.com/path?foo=bar&baz=qux")
	if baseURL != "https://example.com/path" {
		t.Fatalf("baseURL mismatch: got %q", baseURL)
	}
	if params["foo"] != "bar" || params["baz"] != "qux" || len(params) != 2 {
		t.Fatalf("params mismatch: got %#v", params)
	}

	baseURL, params = parser.parseURLAndParams("https://example.com/noquery")
	if baseURL != "https://example.com/noquery" {
		t.Fatalf("baseURL mismatch: got %q", baseURL)
	}
	if len(params) != 0 {
		t.Fatalf("expected empty params, got %#v", params)
	}
}

func TestUnifiedRequestParser_ValidateRequest(t *testing.T) {
	parser := NewUnifiedRequestParser()

	if err := parser.ValidateRequest(nil); err == nil {
		t.Fatalf("expected error for nil request")
	}

	if err := parser.ValidateRequest(&models.ParsedRequest{Method: "GET", URL: "example.com"}); err == nil {
		t.Fatalf("expected error for URL without scheme")
	}

	if err := parser.ValidateRequest(&models.ParsedRequest{Method: "FETCH", URL: "https://example.com"}); err == nil {
		t.Fatalf("expected error for invalid method")
	}

	if err := parser.ValidateRequest(&models.ParsedRequest{Method: "POST", URL: "https://example.com"}); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

func TestUnifiedRequestParser_GeneratePythonCode_UsesBaseURLAndParams(t *testing.T) {
	parser := NewUnifiedRequestParser()

	code := parser.GeneratePythonCode(&models.ParsedRequest{
		Method: "GET",
		URL:    "https://example.com/api?foo=bar",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	})

	if !strings.Contains(code, "url = \"https://example.com/api\"") {
		t.Fatalf("expected code to contain base url, got:\n%s", code)
	}
	if !strings.Contains(code, "params = {") || !strings.Contains(code, "\"foo\": \"bar\"") {
		t.Fatalf("expected code to contain params, got:\n%s", code)
	}
	if !strings.Contains(code, "headers = {") || !strings.Contains(code, "\"Accept\": \"application/json\"") {
		t.Fatalf("expected code to contain headers, got:\n%s", code)
	}
	if !strings.Contains(code, "response = requests.get(") {
		t.Fatalf("expected code to call requests.get, got:\n%s", code)
	}
}

