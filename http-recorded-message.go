package main

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const recordBoundaryHeader = "x-record-boundary"
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

	reqClosed bool
	req       *http.Request
	res       *http.Response
}

func newHTTPRecordedMessage(r io.Reader) *httpRecordedMessage {
	sc := *bufio.NewScanner(r)
	sc.Split(readCRLF)
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

	var chunkedBody []byte
	reqBoundary, reqBodyReading := "", false
	for hrm.scanner.Scan() {
		data := hrm.scanner.Bytes()

		if reqBoundary != "" && reqBodyReading {
			eof := false
			if string(data) == reqBoundary {
				chunkedBody = chunkedBody[:len(chunkedBody)-2] // remove '\r\n'
				eof = true
			}

			if chunkedBody != nil {
				hrm.scannerWritter.Write([]byte(strconv.FormatInt(int64(len(chunkedBody)), 16)))
				hrm.scannerWritter.Write(crlf)
				hrm.scannerWritter.Write(chunkedBody)
				hrm.scannerWritter.Write(crlf)
			}

			if eof {
				hrm.scannerWritter.Write([]byte{'0', '\r', '\n', '\r', '\n'})
				reqBoundary, reqBodyReading = "", false
			} else {
				chunkedBody = data
			}
			continue
		}

		if reqBoundary != "" && bytes.Compare(data, crlf) == 0 {
			hrm.scannerWritter.Write([]byte("Transfer-Encoding: chunked\r\n"))
			reqBodyReading = true
		}

		if hrm.req == nil && strings.ToLower(string(data[0:len(recordBoundaryHeader)+1])) == recordBoundaryHeader+":" {
			if reqBoundary != "" {
				hl := recordBoundaryHeader + ": " + reqBoundary[0:len(reqBoundary)-4] + "\r\n"
				hrm.scannerWritter.Write([]byte(hl))
			}
			reqBoundary = strings.TrimSpace(string(data[len(recordBoundaryHeader)+1:])) + "--\r\n"
			continue
		}

		hrm.scannerWritter.Write(data)
	}
}

func (hrm *httpRecordedMessage) Request() (req *http.Request, err error) {
	if hrm.req != nil {
		return hrm.req, nil
	}

	hrm.req, err = http.ReadRequest(bufio.NewReader(hrm.record))
	if hrm.req != nil {
		hrm.req.Body = &readcloserRecorded{ReadCloser: hrm.req.Body}
	}
	return hrm.req, err
}

func (hrm *httpRecordedMessage) Response() (res *http.Response, err error) {
	if hrm.res != nil {
		return hrm.res, nil
	}

	if hrm.req == nil {
		_, err = hrm.Request()
		if err != nil {
			return
		}
	}
	if c, ok := hrm.req.Body.(*readcloserRecorded); ok && !c.closed {
		io.Copy(io.Discard, hrm.req.Body)
	}

	reader := bufio.NewReader(hrm.record)
	for b, err := reader.Peek(2); err == nil && bytes.Compare(b, crlf) == 0; b, err = reader.Peek(2) {
		reader.ReadByte()
		reader.ReadByte()
	}

	hrm.res, err = http.ReadResponse(reader, hrm.req)
	return hrm.res, err
}
