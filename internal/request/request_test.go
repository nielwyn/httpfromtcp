package request

import (
	"io"
	"strings"
	"testing"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func TestRequestLineParse(t *testing.T) {
	// Valid GET with chunked reader (3 bytes per read)
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.RequestLine.Method != "GET" {
		t.Errorf("Method = %q, want %q", r.RequestLine.Method, "GET")
	}
	if r.RequestLine.RequestTarget != "/" {
		t.Errorf("RequestTarget = %q, want %q", r.RequestLine.RequestTarget, "/")
	}
	if r.RequestLine.HttpVersion != "1.1" {
		t.Errorf("HttpVersion = %q, want %q", r.RequestLine.HttpVersion, "1.1")
	}

	// Valid GET with path and chunked reader (1 byte per read)
	reader = &chunkReader{
		data:            "GET /pomegranate HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.RequestLine.Method != "GET" {
		t.Errorf("Method = %q, want %q", r.RequestLine.Method, "GET")
	}
	if r.RequestLine.RequestTarget != "/pomegranate" {
		t.Errorf("RequestTarget = %q, want %q", r.RequestLine.RequestTarget, "/pomegranate")
	}
	if r.RequestLine.HttpVersion != "1.1" {
		t.Errorf("HttpVersion = %q, want %q", r.RequestLine.HttpVersion, "1.1")
	}

	// Valid GET request line
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.RequestLine.Method != "GET" {
		t.Errorf("Method = %q, want %q", r.RequestLine.Method, "GET")
	}
	if r.RequestLine.RequestTarget != "/" {
		t.Errorf("RequestTarget = %q, want %q", r.RequestLine.RequestTarget, "/")
	}
	if r.RequestLine.HttpVersion != "1.1" {
		t.Errorf("HttpVersion = %q, want %q", r.RequestLine.HttpVersion, "1.1")
	}

	// Valid POST request line with path
	r, err = RequestFromReader(strings.NewReader("POST /pomegranate HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.RequestLine.Method != "POST" {
		t.Errorf("Method = %q, want %q", r.RequestLine.Method, "POST")
	}
	if r.RequestLine.RequestTarget != "/pomegranate" {
		t.Errorf("RequestTarget = %q, want %q", r.RequestLine.RequestTarget, "/pomegranate")
	}
	if r.RequestLine.HttpVersion != "1.1" {
		t.Errorf("HttpVersion = %q, want %q", r.RequestLine.HttpVersion, "1.1")
	}

	// Invalid request line — method and target out of order
	r, err = RequestFromReader(strings.NewReader("/ GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if r != nil {
		t.Errorf("request = %v, want nil", r)
	}

	// Invalid request line — unsupported HTTP version
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/2.9\r\nHost: localhost:42069\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if r != nil {
		t.Errorf("request = %v, want nil", r)
	}

	// Invalid request line — missing method (target as first token)
	r, err = RequestFromReader(strings.NewReader("/pomegranate HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if r != nil {
		t.Errorf("request = %v, want nil", r)
	}

	// Invalid request line — missing target
	r, err = RequestFromReader(strings.NewReader("GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if r != nil {
		t.Errorf("request = %v, want nil", r)
	}
}

func TestHeadersParse(t *testing.T) {
	// Valid standard headers with chunked reader
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.Headers["host"] != "localhost:42069" {
		t.Errorf("Headers[\"host\"] = %q, want %q", r.Headers["host"], "localhost:42069")
	}
	if r.Headers["user-agent"] != "curl/7.81.0" {
		t.Errorf("Headers[\"user-agent\"] = %q, want %q", r.Headers["user-agent"], "curl/7.81.0")
	}
	if r.Headers["accept"] != "*/*" {
		t.Errorf("Headers[\"accept\"] = %q, want %q", r.Headers["accept"], "*/*")
	}

	// Valid empty headers — no headers between request line and body
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\n\r\n"))
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if len(r.Headers) != 0 {
		t.Errorf("Headers = %v, want empty", r.Headers)
	}

	// Valid duplicate headers — values combined with comma
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n"))
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if want := "text/html, application/json"; r.Headers["accept"] != want {
		t.Errorf("Headers[\"accept\"] = %q, want %q", r.Headers["accept"], want)
	}

	// Valid case insensitive headers — keys stored as lowercase
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nContent-Type: application/json\r\n\r\n"))
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if r.Headers["content-type"] != "application/json" {
		t.Errorf("Headers[\"content-type\"] = %q, want %q", r.Headers["content-type"], "application/json")
	}
	if r.Headers["Content-Type"] != "" {
		t.Errorf("Headers[\"Content-Type\"] = %q, want empty", r.Headers["Content-Type"])
	}

	// Invalid missing end of headers — no final CRLF
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Invalid header — missing colon separator
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBodyParse(t *testing.T) {
	// Valid standard body with chunked reader
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}
	if got, want := string(r.Body), "hello world!\n"; got != want {
		t.Errorf("Body = %q, want %q", got, want)
	}

	// Valid empty body with zero content-length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}

	// Valid empty body with no content-length header
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}

	// Valid body without content-length header — body ignored
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	if err != nil {
		t.Fatalf("RequestFromReader error: %v", err)
	}
	if r == nil {
		t.Fatal("request is nil")
	}

	// Invalid body shorter than content-length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from
// a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	endIndex = min(endIndex, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}
