package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Valid GET with path and chunked reader (1 byte per read)
	reader = &chunkReader{
		data:            "GET /pomegranate HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/pomegranate", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Valid GET request line
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Valid POST request line with path
	r, err = RequestFromReader(strings.NewReader("POST /pomegranate HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/pomegranate", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Invalid request line — method and target out of order
	r, err = RequestFromReader(strings.NewReader("/ GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Invalid request line — unsupported HTTP version
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/2.9\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Invalid request line — missing method (target as first token)
	r, err = RequestFromReader(strings.NewReader("/pomegranate HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Invalid request line — missing target
	r, err = RequestFromReader(strings.NewReader("GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)
}

func TestHeadersParse(t *testing.T) {
	// Valid headers parsed correctly with chunked reader
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Valid empty headers — no headers between request line and body
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers)

	// Valid duplicate headers — values combined with comma
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "text/html, application/json", r.Headers["accept"])

	// Valid case insensitive headers — keys stored as lowercase
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nContent-Type: application/json\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "application/json", r.Headers["content-type"])
	assert.Equal(t, "", r.Headers["Content-Type"])

	// Invalid missing end of headers — no final CRLF
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\n"))
	require.Error(t, err)

	// Invalid header — missing colon separator
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
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
