package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
)

type corazaWaf struct {
	*coraza.Waf
}

func newCorazaWaf() (cw corazaWaf, err error) {
	waf := coraza.NewWaf()
	parser, _ := seclang.NewParser(waf)
	for _, f := range []string{"crs/coraza.conf", "crs/crs-setup.conf", "crs/rules/*.conf"} {
		if err := parser.FromFile(f); err != nil {
			return cw, fmt.Errorf("error reading rules file from %s: %w", f, err)
		}
	}
	return corazaWaf{waf}, nil
}

type truncatedBuffer struct {
	bytes.Buffer
	limit int64
}

func (lb *truncatedBuffer) Write(p []byte) (n int, err error) {
	if lb.limit <= 0 {
		return len(p), nil
	}

	l := len(p)
	if lb.limit < int64(l) {
		l = int(lb.limit)
	}
	n, err = lb.Buffer.Write(p[:l])
	if err != nil {
		return n, err
	}
	lb.limit -= int64(n)

	return len(p) - (l - n), nil
}

func (cw corazaWaf) ProcessRecord(record io.Reader) (sc *scores, err error) {
	tx := cw.NewTransaction()
	defer func() {
		tx.ProcessLogging()
		tx.Clean()
	}()

	reader := bufio.NewReader(record)

	// request
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, fmt.Errorf("error parsing request: %w", err)
	}
	defer req.Body.Close()

	// process connection
	raddr := req.RemoteAddr
	if raddr == "" {
		raddr = req.Header.Get("x-forwarded-for")
	}
	client, port := "", 0
	spl := strings.Split(raddr, ":")
	if len(spl) > 1 {
		client = strings.Join(spl[0:len(spl)-1], "")
		port, _ = strconv.Atoi(spl[len(spl)-1])
	}
	tx.ProcessConnection(client, port, "", 0)

	// process uri
	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)

	// process header
	for k, vr := range req.Header {
		for _, v := range vr {
			tx.AddRequestHeader(k, v)
		}
	}
	if req.Host != "" {
		tx.AddRequestHeader("Host", req.Host)
	}
	tx.ProcessRequestHeaders()

	// process body
	reqBody := truncatedBuffer{limit: tx.RequestBodyLimit}
	mw := io.MultiWriter(&reqBody, tx.RequestBodyBuffer)
	if _, err = io.Copy(mw, req.Body); err != nil {
		return nil, fmt.Errorf("error copying request bode: %w", err)
	}
	if _, err := tx.ProcessRequestBody(); err != nil {
		return nil, fmt.Errorf("error processing request: %w", err)
	}

	// skip any CRLF
	for b, err := reader.Peek(2); err == nil && b[0] == '\r' && b[1] == '\n'; b, err = reader.Peek(2) {
		reader.ReadByte()
		reader.ReadByte()
	}

	// remove content-encoding && transfer-encoding
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// default max header line 16kb
		if len(data) > 16384 {
			return len(data), data, nil
		}

		if i := bytes.IndexByte(data, '\n'); i >= 1 && data[i-1] == '\r' {
			return i + 1, data[:i+1], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	})

	resr, resw := io.Pipe()
	go func() {
		defer resw.Close()

		body := false
		for scanner.Scan() {
			chunk := scanner.Bytes()
			if len(chunk) == 2 && chunk[0] == '\r' && chunk[1] == '\n' {
				body = true
			}

			if h := strings.ToLower(string(chunk)); !body &&
				(strings.HasPrefix(h, "transfer-encoding:") || strings.HasPrefix(h, "content-encoding:")) {
				continue
			}

			if _, err := resw.Write(chunk); err != nil {
				log.Printf("error filtering raw response data : %v", err)
				return
			}
		}
	}()

	// response
	res, err := http.ReadResponse(bufio.NewReader(resr), req)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	defer res.Body.Close()

	// response header
	for k, v := range res.Header {
		tx.AddResponseHeader(k, strings.Join(v, ","))
	}
	tx.ProcessResponseHeaders(res.StatusCode, res.Proto)

	// response body
	resBody := truncatedBuffer{limit: tx.ResponseBodyLimit}
	mw = io.MultiWriter(&resBody, tx.ResponseBodyBuffer)
	if _, err := io.Copy(mw, res.Body); err != nil || !tx.IsProcessableResponseBody() {
		return nil, fmt.Errorf("error copying response body: %w", err)
	}
	if _, err := tx.ProcessResponseBody(); err != nil {
		return nil, fmt.Errorf("error processing response body: %w", err)
	}

	return newScore(tx), nil
}
