package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Valid sigle header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069     \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nAccept: text/html, application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 23, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "text/html, application/json", headers["accept"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Valid header with uppercase key — map key must be lowercased
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, "", headers["Content-Type"])
	assert.False(t, done)

	// Valid duplicate header key — values are combined with comma
	headers = NewHeaders()
	data = []byte("Accept: text/html\r\nAccept: application/json\r\nAccept: application/xml\r\n\r\n")
	offset := 0
	n, done, err = headers.Parse(data[offset:])
	require.NoError(t, err)
	assert.False(t, done)
	offset += n
	n, done, err = headers.Parse(data[offset:])
	require.NoError(t, err)
	assert.False(t, done)
	offset += n
	n, done, err = headers.Parse(data[offset:])
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "text/html, application/json, application/xml", headers["accept"])

	// Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Invalid missing colon
	headers = NewHeaders()
	data = []byte("InvalidHeader\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Invalid character in field name (@)
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Invalid non-ASCII character in field name (©)
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

}

func NewHeaders() Headers {
	return make(Headers)
}
