package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const transferEncodingHeader = "transfer-encoding"
const maxHeaderLine = 16384

var crlf = []byte{'\r', '\n'}

func readCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, crlf); i >= 0 {
		return i + 2, data[:i+2], nil
	}

	if len(data) > maxHeaderLine || atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func skipLeadingCRLF(r *bufio.Reader) *bufio.Reader {
	for b, err := r.Peek(2); err == nil && bytes.Compare(crlf, b) == 0; b, err = r.Peek(2) {
		r.ReadByte()
		r.ReadByte()
	}
	return r
}

type startEnd struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end"`
}
type httpRecordedMessageContextDuration struct {
	Proxy *startEnd `json:"proxy"`
	Total startEnd  `json:"total"`
}
type httpRecordedMessageContextCredential struct {
	Username    string `json:"username"`
	PublicKey   string `json:"publicKey"`
	Fingerprint string `json:"fingerprint"`
}
type httpRecordedMessageContextDomain struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type httpRecordedMessageContext struct {
	ID string `json:"id"`

	Duration   httpRecordedMessageContextDuration   `json:"duration"`
	Credential httpRecordedMessageContextCredential `json:"credential"`
	Domain     httpRecordedMessageContextDomain     `json:"domain"`
}

type readcloserRecorded struct {
	io.ReadCloser

	closed bool
}

func (rc *readcloserRecorded) Close() error {
	rc.closed = true
	return rc.ReadCloser.Close()
}

type httpRecordedMessage struct {
	scanner        bufio.Scanner
	scannerWritter *io.PipeWriter
	record         io.Reader

	req *http.Request
	res *http.Response
	ctx *httpRecordedMessageContext
}

func newHTTPRecordedMessage(r io.Reader) *httpRecordedMessage {
	sc := *bufio.NewScanner(r)
	r, w := io.Pipe()
	hrm := &httpRecordedMessage{
		scanner:        sc,
		scannerWritter: w,
		record:         r,
	}
	go hrm.feed()

	return hrm
}

func (hrm *httpRecordedMessage) feed() {
	defer hrm.scannerWritter.Close()

	var eofLine []byte
	var body []byte
	bodyReading := false
	chunked := false

	skipEmpty := func() []byte {
		for hrm.scanner.Scan() {
			data := hrm.scanner.Bytes()
			if bytes.Compare(crlf, data) == 0 {
				continue
			}
			return data
		}
		return nil
	}

	writeBody := func(body []byte, chunked bool) (n int, err error) {
		if !chunked {
			return hrm.scannerWritter.Write(body)
		}

		size := []byte(strconv.FormatInt(int64(len(body)), 16))
		for _, b := range [][]byte{size, crlf, body, crlf} {
			i, err := hrm.scannerWritter.Write(b)
			n += i
			if err != nil {
				return n, err
			}
		}
		return
	}

	hrm.scanner.Split(readCRLF)
	for hrm.scanner.Scan() {
		data := hrm.scanner.Bytes()

		if eofLine == nil {
			eofLine = []byte(string(data))
			data = skipEmpty()
		}

		if bodyReading {
			eof := false
			if bytes.Compare(data, eofLine) == 0 {
				body = body[:len(body)-2] // remove '\r\n'
				eof = true
			}

			if body != nil {
				writeBody(body, chunked)
			}

			if !eof {
				body = data
				continue
			}

			if chunked {
				hrm.scannerWritter.Write([]byte{'0', '\r', '\n', '\r', '\n'})
			}
			body, chunked, bodyReading = nil, false, false
			data = skipEmpty()
		}

		if !bodyReading && bytes.Compare(data, crlf) == 0 {
			bodyReading = true
		}

		if h := string(data[:len(transferEncodingHeader)+1]); strings.EqualFold(transferEncodingHeader+":", h) {
			hv := string(data[len(transferEncodingHeader)+1:])
			for _, v := range strings.Split(hv, ",") {
				if strings.EqualFold("chunked", strings.TrimSpace(v)) {
					chunked = true
					break
				}
			}
		}

		hrm.scannerWritter.Write(data)
	}
}

func (hrm *httpRecordedMessage) Request() (req *http.Request, err error) {
	if hrm.req != nil {
		return hrm.req, nil
	}

	r := skipLeadingCRLF(bufio.NewReader(hrm.record))
	hrm.req, err = http.ReadRequest(r)
	return hrm.req, err
}

func (hrm *httpRecordedMessage) Response() (res *http.Response, err error) {
	if hrm.res != nil {
		return hrm.res, nil
	}

	if hrm.req == nil {
		if _, err := hrm.Request(); err != nil {
			return nil, err
		}
	}
	io.Copy(io.Discard, hrm.req.Body)

	r := skipLeadingCRLF(bufio.NewReader(hrm.record))
	hrm.res, err = http.ReadResponse(r, hrm.req)
	return hrm.res, err
}

func (hrm *httpRecordedMessage) Context() (ctx *httpRecordedMessageContext, err error) {
	if hrm.ctx != nil {
		return hrm.ctx, nil
	}

	if hrm.res == nil {
		if _, err := hrm.Response(); err != nil {
			return nil, err
		}
	}
	io.Copy(io.Discard, hrm.res.Body)

	r := skipLeadingCRLF(bufio.NewReader(hrm.record))
	err = json.NewDecoder(r).Decode(&hrm.ctx)
	return hrm.ctx, nil
}
