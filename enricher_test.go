package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScorer(t *testing.T) {
	sc, err := newEnricher()
	require.Nil(t, err, "unexpected error in instantiating scorer")

	f, err := os.ReadFile("testdata/record.txt")
	require.Nil(t, err, "unexpected error in reading test data")

	s, err := sc.ProcessRecord(bytes.NewReader(f))
	assert.NoError(t, err, "should not return error")
	assert.NotNil(t, s, "should produce non nil score")
}
