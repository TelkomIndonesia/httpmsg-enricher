package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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

	if len(data) > maxHeaderLine || atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

type httpRecordedMessage struct {
	scanner        bufio.Scanner
	scannerWritter *io.PipeWriter
	record         io.Reader

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
		record:         r,
	}
	go hrm.feed()

	return hrm
}

func (hrm *httpRecordedMessage) feed() {
	defer hrm.scannerWritter.Close()

	var body, eofLine []byte
	var bodyReading, chunked bool
	hrm.scanner.Split(splitCRLF)
	for hrm.scanner.Scan() {
		data := hrm.scanner.Bytes()

		if eofLine == nil {
			eofLine = []byte(string(data))
			data = hrm.discardEmpty()
		}

		if bodyReading {
			eof := false
			if bytes.Compare(data, eofLine) == 0 && bytes.Compare(body[len(body)-2:], crlf) == 0 {
				body = body[:len(body)-2] // remove '\r\n'
				eof = true
			}

			hrm.feedBody(body, chunked)

			if !eof {
				body = data
				continue
			}

			if chunked {
				hrm.scannerWritter.Write([]byte{'0', '\r', '\n', '\r', '\n'})
			}
			body, chunked, bodyReading = nil, false, false
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
		return data
	}
	return nil
}

func (hrm *httpRecordedMessage) feedBody(body []byte, chunked bool) (n int, err error) {
	if body == nil {
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

	r := bufio.NewReader(hrm.record)
	hrm.req, err = http.ReadRequest(r)
	return hrm.req, err
}

func (hrm *httpRecordedMessage) Response() (_ *http.Response, err error) {
	if hrm.res != nil {
		return hrm.res, nil
	}

	if hrm.req == nil {
		if _, err := hrm.Request(); err != nil {
			return nil, err
		}
	}
	io.Copy(io.Discard, hrm.req.Body)

	r := bufio.NewReader(hrm.record)
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

	r := bufio.NewReader(hrm.record)
	err = json.NewDecoder(r).Decode(&hrm.ctx)
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
