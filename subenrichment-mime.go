package main

import (
	"io"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

type writableMimeReader struct {
	r    *io.PipeReader
	w    *io.PipeWriter
	mime string
}

func newWritableMimeReader() *writableMimeReader {
	r, w := io.Pipe()
	me := &writableMimeReader{
		r: r,
		w: w,
	}
	go me.detectMime()
	return me
}

func (me *writableMimeReader) detectMime() {
	mtype, err := mimetype.DetectReader(me.r)
	if err != nil {
		return
	}
	me.mime = mtype.String()

	io.Copy(io.Discard, me.r)
}

func (me *writableMimeReader) Close() error {
	return me.w.Close()
}

var _ subEnrichment = &mimeEnrichment{}

type mimeEnrichment struct {
	req *writableMimeReader
	res *writableMimeReader
}

func (erc *mimeEnrichment) Close() error {
	erc.req.Close()
	erc.res.Close()
	return nil
}

func (erc *mimeEnrichment) requestBodyWriter() closableWriter {
	return erc.req.w
}
func (erc *mimeEnrichment) processRequest(req *http.Request) (err error) { return nil }

func (erc *mimeEnrichment) responseBodyWriter() closableWriter {
	return erc.res.w
}
func (erc *mimeEnrichment) processResponse(res *http.Response) (err error) { return nil }

func (erc *mimeEnrichment) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
	doc.HTTP.Request.MimeType = erc.req.mime
	doc.HTTP.Response.MimeType = erc.res.mime
	return
}
