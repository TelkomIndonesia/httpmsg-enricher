package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
)

type corazaWaf struct {
	waf *coraza.Waf
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
	tx := cw.waf.NewTransaction()
	defer func() {
		tx.ProcessLogging()
		tx.Clean()
	}()

	msg := newHTTPRecordedMessage(record)

	// request
	req, err := msg.Request()
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

	// response
	res, err := msg.Response()
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
