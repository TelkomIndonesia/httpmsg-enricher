package main

import (
	"io"
	"net/http"

	ecsx "github.com/telkomindonesia/httpmsg-enricher/ecs/custom"
)

type subEnricher interface {
	requestBodyWriter() io.WriteCloser
	processRequest(req *http.Request) (err error)
	responseBodyWriter() io.WriteCloser
	processResponse(res *http.Response) (err error)
	enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error)

	Close() error
}
