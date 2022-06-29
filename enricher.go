package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
)

type enricher struct {
	waf *coraza.Waf
}

func newEnricher() (cw enricher, err error) {
	waf := coraza.NewWaf()
	parser, _ := seclang.NewParser(waf)
	for _, f := range []string{"crs/coraza.conf", "crs/crs-setup.conf", "crs/rules/*.conf"} {
		if err := parser.FromFile(f); err != nil {
			return cw, fmt.Errorf("error reading rules file from %s: %w", f, err)
		}
	}
	return enricher{waf}, nil
}

func (er enricher) ProcessRequest(tx *coraza.Transaction, req *http.Request, newBody io.ReadWriteCloser) (err error) {
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
	mw := io.MultiWriter(newBody, tx.RequestBodyBuffer)
	if _, err = io.Copy(mw, req.Body); err != nil {
		return fmt.Errorf("error copying request bode: %w", err)
	}
	if _, err := tx.ProcessRequestBody(); err != nil {
		return fmt.Errorf("error processing request: %w", err)
	}

	req.Body = newBody
	return
}

func (er enricher) ProcessResponse(tx *coraza.Transaction, res *http.Response, newBody io.ReadWriteCloser) (err error) {
	// response header
	for k, v := range res.Header {
		tx.AddResponseHeader(k, strings.Join(v, ","))
	}
	tx.ProcessResponseHeaders(res.StatusCode, res.Proto)

	// response body
	mw := io.MultiWriter(newBody, tx.ResponseBodyBuffer)
	if _, err := io.Copy(mw, res.Body); err != nil || !tx.IsProcessableResponseBody() {
		return fmt.Errorf("error copying response body: %w", err)
	}
	if _, err := tx.ProcessResponseBody(); err != nil {
		return fmt.Errorf("error processing response body: %w", err)
	}
	res.Body = newBody

	return
}

func (er enricher) ProcessRecord(record io.Reader) (sc *scores, err error) {
	tx := er.waf.NewTransaction()
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
	err = er.ProcessRequest(tx, req, newTruncatedBuffer(tx.RequestBodyLimit))
	if err != nil {
		return nil, fmt.Errorf("Error processing request: %w", err)
	}

	// response
	res, err := msg.Response()
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	defer res.Body.Close()
	err = er.ProcessResponse(tx, res, newTruncatedBuffer(tx.ResponseBodyLimit))
	if err != nil {
		return nil, fmt.Errorf("Error processing response: %w", err)
	}

	return newScores(tx), nil
}
