package parser

import "testing"

func TestCurlRequestParser_ParseDefaultsToPostWhenBodyProvided(t *testing.T) {
	parser := NewCurlRequestParser()

	req, err := parser.Parse(`curl 'https://example.com/login' -H 'Content-Type: application/json' --data-raw '{"username":"alice","password":"secret"}'`)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}

	if req.Method != "POST" {
		t.Fatalf("expected method POST, got %q", req.Method)
	}

	if req.URL != "https://example.com/login" {
		t.Fatalf("expected URL preserved, got %q", req.URL)
	}

	if req.Body != `{"username":"alice","password":"secret"}` {
		t.Fatalf("expected body preserved, got %q", req.Body)
	}
}

func TestCurlRequestParser_ExtractURLSkipsOptionValues(t *testing.T) {
	parser := NewCurlRequestParser()

	req, err := parser.Parse(`curl -H 'Accept: application/json' --data-raw '{"a":1}' 'https://example.com/api/test'`)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}

	if req.URL != "https://example.com/api/test" {
		t.Fatalf("expected URL to be detected from trailing argument, got %q", req.URL)
	}
}
