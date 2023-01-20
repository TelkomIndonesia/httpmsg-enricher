package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randStringBytes(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestReadCRLF(t *testing.T) {
	longstr := randStringBytes(20000)

	table := []struct {
		input string
		exp   []string
	}{
		{
			input: "asda\nsd\rsdasd\r\nasdasadasdasdas\r\nda",
			exp:   []string{"asda\nsd\rsdasd\r\n", "asdasadasdasdas\r\n", "da"},
		},
		{
			input: longstr + "\r\nkey: value\r\n",
			exp:   []string{longstr + "\r\n", "key: value\r\n"},
		},
	}

	for i, tt := range table {
		scanner := bufio.NewScanner(strings.NewReader(tt.input))
		scanner.Split(splitCRLF)

		got := []string{}
		for scanner.Scan() {
			got = append(got, scanner.Text())
			assert.Nilf(t, scanner.Err(), "should not produce error. index %d", i)
		}
		assert.ElementsMatchf(t, got, tt.exp, "should correctly split strings. index %d", i)
	}
}

func TestRecordedMessage(t *testing.T) {
	table := []struct {
		file string
	}{
		{file: "testdata/record.txt"},
		{file: "testdata/record1.txt"},
		{file: "testdata/record2.txt"},
		{file: "testdata/record3.txt"},
		{file: "testdata/record4.txt"},
	}

	for _, tt := range table {

		f, err := os.Open(tt.file)
		require.Nil(t, err, "unexpected error in reading test data")

		h := newHTTPRecordedMessage(io.ReadCloser(f))
		req, err := h.Request()
		require.Nil(t, err, "should not return error")
		require.NotNil(t, req, "should not return nil request")
		b, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err, "req body should be readable")
		t.Log(len(b))

		res, err := h.Response()
		require.Nil(t, err, "should not return error")
		require.NotNil(t, res, "should not return nil response")
		b, err = ioutil.ReadAll(res.Body)
		require.Nil(t, err, "res body should be readable")
		t.Log(len(b))

		ctx, err := h.Context()
		require.Nil(t, err, "should not return error")
		require.NotNil(t, ctx, "should not return nil context")
		b, err = json.Marshal(ctx)
		require.Nil(t, err, "res body should be marshallable")
		t.Log(string(b))
	}

}
