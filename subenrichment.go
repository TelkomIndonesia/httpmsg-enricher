package main

import (
	"io"
	"net/http"

	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

type subEnrichment interface {
	requestBodyWriter() io.WriteCloser
	processRequest(req *http.Request) (err error)
	responseBodyWriter() io.WriteCloser
	processResponse(res *http.Response) (err error)
	enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error)

	Close() error
}
