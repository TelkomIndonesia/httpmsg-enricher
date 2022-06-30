package main

import (
	"bytes"
	"encoding/json"
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

	s, err := sc.EnrichRecord(bytes.NewReader(f))
	defer s.Close()
	assert.NoError(t, err, "should not return error")
	assert.NotNil(t, s, "should produce non nil score")
	ecs, _ := s.toECS()
	b, _ := json.Marshal(ecs)
	t.Log(string(b))
}
