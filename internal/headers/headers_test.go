package headers

import (
	"testing"
)

func TestHeaders(t *testing.T) {
	// Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if headers["host"] != "localhost:42069" {
		t.Errorf("headers[\"host\"] = %q, want %q", headers["host"], "localhost:42069")
	}
	if n != 23 {
		t.Errorf("n = %d, want 23", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Valid sigle header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069     \r\n\r\n")
	n, done, err = headers.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if headers["host"] != "localhost:42069" {
		t.Errorf("headers[\"host\"] = %q, want %q", headers["host"], "localhost:42069")
	}
	if n != 32 {
		t.Errorf("n = %d, want 32", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nAccept: text/html, application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if n != 23 {
		t.Errorf("n = %d, want 23", n)
	}
	if done {
		t.Error("done = true, want false")
	}
	n, done, err = headers.Parse(data[n:])
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if headers["host"] != "localhost:42069" {
		t.Errorf("headers[\"host\"] = %q, want %q", headers["host"], "localhost:42069")
	}
	if headers["accept"] != "text/html, application/json" {
		t.Errorf("headers[\"accept\"] = %q, want %q", headers["accept"], "text/html, application/json")
	}
	if n != 37 {
		t.Errorf("n = %d, want 37", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Valid header with uppercase key — map key must be lowercased
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if headers["content-type"] != "application/json" {
		t.Errorf("headers[\"content-type\"] = %q, want %q", headers["content-type"], "application/json")
	}
	if headers["Content-Type"] != "" {
		t.Errorf("headers[\"Content-Type\"] = %q, want empty", headers["Content-Type"])
	}
	if done {
		t.Error("done = true, want false")
	}

	// Valid duplicate header key — values are combined with comma
	headers = NewHeaders()
	data = []byte("Accept: text/html\r\nAccept: application/json\r\nAccept: application/xml\r\n\r\n")
	offset := 0
	for range 3 {
		n, done, err = headers.Parse(data[offset:])
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if done {
			t.Error("done = true, want false")
		}
		offset += n
	}
	if want := "text/html, application/json, application/xml"; headers["accept"] != want {
		t.Errorf("headers[\"accept\"] = %q, want %q", headers["accept"], want)
	}

	// Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if n != 2 {
		t.Errorf("n = %d, want 2", n)
	}
	if !done {
		t.Error("done = false, want true")
	}

	// Invalid missing colon
	headers = NewHeaders()
	data = []byte("InvalidHeader\r\n\r\n")
	n, done, err = headers.Parse(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Invalid character in field name (@)
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
	if done {
		t.Error("done = true, want false")
	}

	// Invalid non-ASCII character in field name (©)
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n != 0 {
		t.Errorf("n = %d, want 0", n)
	}
	if done {
		t.Error("done = true, want false")
	}
}

func NewHeaders() Headers {
	return make(Headers)
}
