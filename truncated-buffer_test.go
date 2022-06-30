package main

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncatedBuffer(t *testing.T) {
	in := "12345678901234567890"
	limit := 6
	buff := newTruncatedBuffer(limit)
	i, err := io.Copy(buff, strings.NewReader(in))
	assert.NoError(t, err, "should not return error")
	assert.Equal(t, int64(len(in)), i, "should signal that it writes all byte")
	assert.Equal(t, len(in), buff.Len(), "should return the original number of bytes to be written")
	assert.Equal(t, in[:limit], buff.String(), "should only contain truncated byte")
}
