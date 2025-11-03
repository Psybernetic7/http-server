package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func Test_ValidSingleHeaderWithWhitespace(t *testing.T) {
	h := NewHeaders()
	data := []byte("   Host:    localhost:42069     \r\n\r\n")

	n, done, err := h.Parse(data)

	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "localhost:42069", h["Host"])
	// consumed through the first CRLF only
	assert.Greater(t, n, 0)
}

func Test_ValidTwoHeadersWithExisting(t *testing.T) {
	h := NewHeaders()
	h["Existing"] = "keepme"

	// parse first
	data := []byte("Host: a\r\nContent-Type: text/plain\r\n\r\n")
	n1, done1, err1 := h.Parse(data)
	require.NoError(t, err1)
	assert.False(t, done1)
	assert.Equal(t, "a", h["Host"])

	// parse second (advance by n1)
	n2, done2, err2 := h.Parse(data[n1:])
	require.NoError(t, err2)
	assert.False(t, done2)
	assert.Equal(t, "text/plain", h["Content-Type"])

	// existing remains
	assert.Equal(t, "keepme", h["Existing"])

	// now “done” line
	n3, done3, err3 := h.Parse(data[n1+n2:])
	require.NoError(t, err3)
	assert.True(t, done3)
	assert.Equal(t, 2, n3) // "\r\n"
}

func Test_ValidDone(t *testing.T) {
	h := NewHeaders()
	data := []byte("\r\nrest")

	n, done, err := h.Parse(data)

	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 2, n)
}
