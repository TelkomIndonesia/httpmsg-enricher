package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

const transferEncodingHeader = "transfer-encoding"
const maxHeaderLine = 16384

var crlf = []byte{'\r', '\n'}

func splitCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, crlf); i >= 0 {
		return i + 2, data[:i+2], nil
	}
	if len(data) > maxHeaderLine {
		return maxHeaderLine, data[:maxHeaderLine], nil
	}
	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

type httpRecordedMessage struct {
	scanner        bufio.Scanner
	scannerWritter *io.PipeWriter
	record         *bufio.Reader

	req *http.Request
	res *http.Response
	ctx *httpRecordedMessageContext

	closed bool
}

func newHTTPRecordedMessage(r io.Reader) *httpRecordedMessage {
	sc := *bufio.NewScanner(r)
	r, w := io.Pipe()
	hrm := &httpRecordedMessage{
		scanner:        sc,
		scannerWritter: w,
		record:         bufio.NewReader(r),
	}
	go hrm.feed()

	return hrm
}

func (hrm *httpRecordedMessage) feed() {
	defer hrm.scannerWritter.Close()

	var body, eofLine []byte
	var bodyWritten int
	var bodyReading, chunked bool
	hrm.scanner.Split(splitCRLF)
	for hrm.scanner.Scan() {
		data := append([]byte(nil), hrm.scanner.Bytes()...)

		if eofLine == nil {
			eofLine = data
			data = hrm.discardEmpty()
		}

		if bodyReading {
			eof := false
			if bytes.Compare(data, eofLine) == 0 && bytes.Compare(body[len(body)-2:], crlf) == 0 {
				body = body[:len(body)-2] // remove '\r\n'
				eof = true
			}

			n, _ := hrm.feedBody(body, chunked)
			bodyWritten += n

			if !eof {
				body = data
				continue
			}

			if chunked && bodyWritten > 0 {
				hrm.scannerWritter.Write([]byte{'0', '\r', '\n', '\r', '\n'})
			}
			body, bodyWritten, chunked, bodyReading = nil, 0, false, false
			data = hrm.discardEmpty()
		}

		if isChunked(data) {
			chunked = true

		} else if bytes.Compare(data, crlf) == 0 {
			bodyReading = true
		}

		hrm.scannerWritter.Write(data)
	}
}
func (hrm *httpRecordedMessage) discardEmpty() []byte {
	for hrm.scanner.Scan() {
		data := hrm.scanner.Bytes()
		if bytes.Compare(crlf, data) == 0 {
			continue
		}
		return append([]byte(nil), data...)
	}
	return nil
}

func (hrm *httpRecordedMessage) feedBody(body []byte, chunked bool) (n int, err error) {
	if body == nil || len(body) == 0 {
		return
	}

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

func isChunked(headerline []byte) (chunked bool) {
	l := len(transferEncodingHeader)
	if len(headerline) <= l+1 {
		return
	}

	if h := string(headerline[:l+1]); !strings.EqualFold(transferEncodingHeader+":", h) {
		return
	}

	v := string(headerline[l+1:])
	for _, v := range strings.Split(v, ",") {
		if strings.EqualFold("chunked", strings.TrimSpace(v)) {
			return true
		}
	}
	return
}

func (hrm *httpRecordedMessage) Request() (_ *http.Request, err error) {
	if hrm.req != nil {
		return hrm.req, nil
	}

	hrm.req, err = http.ReadRequest(hrm.record)
	return hrm.req, err
}

func (hrm *httpRecordedMessage) Response() (_ *http.Response, err error) {
	if hrm.res != nil {
		return hrm.res, nil
	}
	if hrm.req == nil {
		return nil, fmt.Errorf("consume the request and its whole body first")
	}

	hrm.res, err = http.ReadResponse(hrm.record, hrm.req)
	return hrm.res, err
}

func (hrm *httpRecordedMessage) Context() (ctx *httpRecordedMessageContext, err error) {
	if hrm.ctx != nil {
		return hrm.ctx, nil
	}
	if hrm.res == nil {
		return nil, fmt.Errorf("consume the response and its whole body first")
	}
	if err := json.NewDecoder(hrm.record).Decode(&hrm.ctx); err != nil && err != io.EOF {
		return nil, err
	}
	return hrm.ctx, nil
}

func (hrm *httpRecordedMessage) Close() (err error) {
	io.Copy(io.Discard, hrm.record)
	if errt := hrm.req.Body.Close(); errt != nil {
		err = multierror.Append(errt)
	}
	if errt := hrm.res.Body.Close(); errt != nil {
		err = multierror.Append(errt)
	}
	return
}
