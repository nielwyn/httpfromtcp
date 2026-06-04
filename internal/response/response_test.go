package response

import (
	"bytes"
	"httpfromtcp/internal/headers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWriter() (*Writer, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return NewWriter(buf), buf
}

func TestWriteStatusLine(t *testing.T) {
	w, buf := newWriter()
	err := w.WriteStatusLine(StatusCodeOK)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 200 OK\r\n", buf.String())

	w, buf = newWriter()
	err = w.WriteStatusLine(StatusCodeBadRequest)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 400 Bad Request\r\n", buf.String())

	w, buf = newWriter()
	err = w.WriteStatusLine(StatusCodeInternalServerError)
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 500 Internal Server Error\r\n", buf.String())
}

func TestWriteStatusLineInvalidState(t *testing.T) {
	w, _ := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	err := w.WriteStatusLine(StatusCodeOK)
	require.Error(t, err)
}

func TestWriteHeaders(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	h := headers.Headers{"content-type": "text/plain", "connection": "close"}
	err := w.WriteHeaders(h)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "content-type: text/plain\r\n")
	assert.Contains(t, buf.String(), "connection: close\r\n")
	assert.Contains(t, buf.String(), "\r\n\r\n")
}

func TestWriteHeadersInvalidState(t *testing.T) {
	w, _ := newWriter()
	err := w.WriteHeaders(headers.Headers{})
	require.Error(t, err)
}

func TestWriteBody(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))
	n, err := w.WriteBody([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Contains(t, buf.String(), "hello")
}

func TestWriteBodyInvalidState(t *testing.T) {
	w, _ := newWriter()
	_, err := w.WriteBody([]byte("hello"))
	require.Error(t, err)
}

func TestWriteChunkedBody(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))

	n, err := w.WriteChunkedBody([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "5\r\nhello\r\n", buf.String()[len("HTTP/1.1 200 OK\r\n\r\n"):])
}

func TestWriteChunkedBodyMultiple(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))

	_, err := w.WriteChunkedBody([]byte("hello"))
	require.NoError(t, err)
	_, err = w.WriteChunkedBody([]byte("world"))
	require.NoError(t, err)

	body := buf.String()[len("HTTP/1.1 200 OK\r\n\r\n"):]
	assert.Equal(t, "5\r\nhello\r\n5\r\nworld\r\n", body)
}

func TestWriteChunkedBodyInvalidState(t *testing.T) {
	w, _ := newWriter()
	_, err := w.WriteChunkedBody([]byte("hello"))
	require.Error(t, err)
}

func TestWriteChunkedBodyDone(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))
	_, err := w.WriteChunkedBody([]byte("hello"))
	require.NoError(t, err)

	_, err = w.WriteChunkedBodyDone()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "0\r\n")
	assert.Equal(t, writerStateTrailers, w.writerState)
}

func TestWriteChunkedBodyDoneInvalidState(t *testing.T) {
	w, _ := newWriter()
	_, err := w.WriteChunkedBodyDone()
	require.Error(t, err)
}

func TestWriteTrailers(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))
	_, err := w.WriteChunkedBody([]byte("hello"))
	require.NoError(t, err)
	_, err = w.WriteChunkedBodyDone()
	require.NoError(t, err)

	err = w.WriteTrailers(headers.Headers{"x-checksum": "abc123"})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "x-checksum: abc123\r\n")
	assert.Contains(t, buf.String(), "\r\n")
	assert.Equal(t, writerStateDone, w.writerState)
}

func TestWriteTrailersEmpty(t *testing.T) {
	w, buf := newWriter()
	require.NoError(t, w.WriteStatusLine(StatusCodeOK))
	require.NoError(t, w.WriteHeaders(headers.Headers{}))
	_, err := w.WriteChunkedBody([]byte("hello"))
	require.NoError(t, err)
	_, err = w.WriteChunkedBodyDone()
	require.NoError(t, err)

	err = w.WriteTrailers(headers.Headers{})
	require.NoError(t, err)
	assert.True(t, len(buf.String()) > 0)
	assert.Equal(t, writerStateDone, w.writerState)
}

func TestWriteTrailersInvalidState(t *testing.T) {
	w, _ := newWriter()
	err := w.WriteTrailers(headers.Headers{})
	require.Error(t, err)
}
