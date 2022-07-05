package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnricher(t *testing.T) {
	table := []struct {
		file string
	}{
		{file: "testdata/record.txt"},
		{file: "testdata/record1.txt"},
	}

	for _, tt := range table {
		sc, err := newEnricher()
		require.Nil(t, err, "unexpected error in instantiating scorer")

		f, err := os.ReadFile(tt.file)
		require.Nil(t, err, "unexpected error in reading test data")

		s, err := sc.EnrichRecord(bytes.NewReader(f))
		require.NoError(t, err, "should not return error")
		require.NotNil(t, s, "should produce non nil score")
		defer s.Close()
		ecs, _ := s.toECS()
		b, _ := json.Marshal(ecs)
		t.Log(string(b))
	}
}
