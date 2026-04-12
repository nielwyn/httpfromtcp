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
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Valid sigle header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:     localhost:42069     \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
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
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, "text/html, application/json", headers["Accept"])
	assert.Equal(t, 41, n)
	assert.False(t, done)

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

}

func NewHeaders() Headers {
	return make(Headers)
}
