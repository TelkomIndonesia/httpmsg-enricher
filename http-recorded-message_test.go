package main

import (
	"bufio"
	"bytes"
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
			input: longstr + "\r\n" + recordBoundaryHeader + ": value\r\n",
			exp:   []string{longstr + "\r\n", recordBoundaryHeader + ": value\r\n"},
		},
	}

	for i, tt := range table {
		scanner := bufio.NewScanner(strings.NewReader(tt.input))
		scanner.Split(readCRLF)

		got := []string{}
		for scanner.Scan() {
			got = append(got, scanner.Text())
			assert.Nilf(t, scanner.Err(), "should not produce error. index %d", i)
		}
		assert.ElementsMatchf(t, got, tt.exp, "should correctly split strings. index %d", i)
	}
}

func TestRecordedMessage(t *testing.T) {
	f, err := os.ReadFile("testdata/record.txt")
	require.Nil(t, err, "unexpected error in reading test data")

	h := newHTTPRecordedMessage(bytes.NewReader(f))
	req, err := h.getRequest()
	require.Nil(t, err, "should not return error")
	require.NotNil(t, req, "should not return nil request")
	b, err := ioutil.ReadAll(req.Body)
	require.Nil(t, err, "req body should be readable")
	t.Log(string(b))

	res, err := h.getResponse()
	require.Nil(t, err, "should not return error")
	require.NotNil(t, res, "should not return nil response")
	b, err = ioutil.ReadAll(res.Body)
	require.Nil(t, err, "res body should be readable")
	t.Log(string(b))
}
