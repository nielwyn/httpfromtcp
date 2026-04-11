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
	// GET Request line with chunkReader
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

	// GET Request line with path chunkReader
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

	// Good GET Request line
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Good POST Request line with path
	r, err = RequestFromReader(strings.NewReader("POST /pomegranate HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/pomegranate", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Invalid method (out of order) request line
	r, err = RequestFromReader(strings.NewReader("/ GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Invalid version in request line
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/2.9\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Invalid method (starts with /)
	r, err = RequestFromReader(strings.NewReader("/pomegranate HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)

	// Missing request target (only 2 parts)
	r, err = RequestFromReader(strings.NewReader("GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)
	require.Nil(t, r)
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
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
