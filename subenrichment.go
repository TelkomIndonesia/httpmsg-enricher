package main

import (
	"net/http"

	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

type subEnrichment interface {
	requestBodyWriter() closableWriter
	processRequest(req *http.Request) (err error)
	responseBodyWriter() closableWriter
	processResponse(res *http.Response) (err error)
	enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error)

	Close() error
}
