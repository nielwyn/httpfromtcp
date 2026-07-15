package response

import (
	"bytes"
	"httpfromtcp/internal/headers"
	"strings"
	"testing"
)

func newWriter() (*Writer, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return NewWriter(buf), buf
}

func TestWriteStatusLine(t *testing.T) {
	w, buf := newWriter()
	err := w.WriteStatusLine(StatusCodeOK)
	if err != nil {
		t.Fatalf("WriteStatusLine(StatusCodeOK) error: %v", err)
	}
	if got, want := buf.String(), "HTTP/1.1 200 OK\r\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	w, buf = newWriter()
	err = w.WriteStatusLine(StatusCodeBadRequest)
	if err != nil {
		t.Fatalf("WriteStatusLine(StatusCodeBadRequest) error: %v", err)
	}
	if got, want := buf.String(), "HTTP/1.1 400 Bad Request\r\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	w, buf = newWriter()
	err = w.WriteStatusLine(StatusCodeInternalServerError)
	if err != nil {
		t.Fatalf("WriteStatusLine(StatusCodeInternalServerError) error: %v", err)
	}
	if got, want := buf.String(), "HTTP/1.1 500 Internal Server Error\r\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	w, buf = newWriter()
	err = w.WriteStatusLine(StatusCode(503))
	if err != nil {
		t.Fatalf("WriteStatusLine(503) error: %v", err)
	}
	if got, want := buf.String(), "HTTP/1.1 503 \r\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteStatusLineInvalidState(t *testing.T) {
	w, _ := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteStatusLine(StatusCodeOK); err == nil {
		t.Fatal("expected error on second WriteStatusLine, got nil")
	}
}

func TestWriteHeaders(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	h := headers.Headers{"content-type": "text/plain", "connection": "close"}
	if err := w.WriteHeaders(h); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}
	for _, want := range []string{"content-type: text/plain\r\n", "connection: close\r\n", "\r\n\r\n"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output %q does not contain %q", buf.String(), want)
		}
	}
}

func TestWriteHeadersInvalidState(t *testing.T) {
	w, _ := newWriter()
	if err := w.WriteHeaders(headers.Headers{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteBody(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}
	n, err := w.WriteBody([]byte("hello"))
	if err != nil {
		t.Fatalf("WriteBody error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteBody returned n = %d, want 5", n)
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("output %q does not contain %q", buf.String(), "hello")
	}
}

func TestWriteBodyInvalidState(t *testing.T) {
	w, _ := newWriter()
	if _, err := w.WriteBody([]byte("hello")); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteChunkedBody(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}

	n, err := w.WriteChunkedBody([]byte("hello"))
	if err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}
	if n != 5 {
		t.Errorf("WriteChunkedBody returned n = %d, want 5", n)
	}
	if got, want := buf.String()[len("HTTP/1.1 200 OK\r\n\r\n"):], "5\r\nhello\r\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestWriteChunkedBodyMultiple(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}

	if _, err := w.WriteChunkedBody([]byte("hello")); err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}
	if _, err := w.WriteChunkedBody([]byte("world")); err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}

	body := buf.String()[len("HTTP/1.1 200 OK\r\n\r\n"):]
	if want := "5\r\nhello\r\n5\r\nworld\r\n"; body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
}

func TestWriteChunkedBodyInvalidState(t *testing.T) {
	w, _ := newWriter()
	if _, err := w.WriteChunkedBody([]byte("hello")); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteChunkedBodyDone(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}
	if _, err := w.WriteChunkedBody([]byte("hello")); err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}

	if _, err := w.WriteChunkedBodyDone(); err != nil {
		t.Fatalf("WriteChunkedBodyDone error: %v", err)
	}
	if !strings.Contains(buf.String(), "0\r\n") {
		t.Errorf("output %q does not contain %q", buf.String(), "0\r\n")
	}
	if w.writerState != writerStateTrailers {
		t.Errorf("writerState = %v, want %v", w.writerState, writerStateTrailers)
	}
}

func TestWriteChunkedBodyDoneInvalidState(t *testing.T) {
	w, _ := newWriter()
	if _, err := w.WriteChunkedBodyDone(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteTrailers(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}
	if _, err := w.WriteChunkedBody([]byte("hello")); err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}
	if _, err := w.WriteChunkedBodyDone(); err != nil {
		t.Fatalf("WriteChunkedBodyDone error: %v", err)
	}

	if err := w.WriteTrailers(headers.Headers{"x-checksum": "abc123"}); err != nil {
		t.Fatalf("WriteTrailers error: %v", err)
	}
	if !strings.Contains(buf.String(), "x-checksum: abc123\r\n") {
		t.Errorf("output %q does not contain %q", buf.String(), "x-checksum: abc123\r\n")
	}
	if !strings.Contains(buf.String(), "\r\n") {
		t.Errorf("output %q does not contain %q", buf.String(), "\r\n")
	}
	if w.writerState != writerStateDone {
		t.Errorf("writerState = %v, want %v", w.writerState, writerStateDone)
	}
}

func TestWriteTrailersEmpty(t *testing.T) {
	w, buf := newWriter()
	if err := w.WriteStatusLine(StatusCodeOK); err != nil {
		t.Fatalf("WriteStatusLine error: %v", err)
	}
	if err := w.WriteHeaders(headers.Headers{}); err != nil {
		t.Fatalf("WriteHeaders error: %v", err)
	}
	if _, err := w.WriteChunkedBody([]byte("hello")); err != nil {
		t.Fatalf("WriteChunkedBody error: %v", err)
	}
	if _, err := w.WriteChunkedBodyDone(); err != nil {
		t.Fatalf("WriteChunkedBodyDone error: %v", err)
	}

	if err := w.WriteTrailers(headers.Headers{}); err != nil {
		t.Fatalf("WriteTrailers error: %v", err)
	}
	if len(buf.String()) == 0 {
		t.Error("expected non-empty output")
	}
	if w.writerState != writerStateDone {
		t.Errorf("writerState = %v, want %v", w.writerState, writerStateDone)
	}
}

func TestWriteTrailersInvalidState(t *testing.T) {
	w, _ := newWriter()
	if err := w.WriteTrailers(headers.Headers{}); err == nil {
		t.Fatal("expected error, got nil")
	}
}
