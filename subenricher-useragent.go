package main

import (
	"fmt"
	"io"
	"net/http"

	ua "github.com/mileusna/useragent"
	"github.com/telkomindonesia/httpmsg-enricher/ecs"
	ecsx "github.com/telkomindonesia/httpmsg-enricher/ecs/custom"
)

var _ subEnricher = &uaEnricher{}

type uaEnricher struct{}

func (uaEnricher) requestBodyWriter() io.WriteCloser              { return nopwc }
func (uaEnricher) processRequest(req *http.Request) (err error)   { return }
func (uaEnricher) responseBodyWriter() io.WriteCloser             { return nopwc }
func (uaEnricher) processResponse(res *http.Response) (err error) { return }
func (uaEnricher) Close() (err error)                             { return }

func (uaEnricher) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
	req, err := msg.Request()
	if err != nil {
		return fmt.Errorf("error getting request: %w", err)
	}

	v := req.Header.Get("user-agent")
	if v == "" {
		return
	}

	uap := ua.Parse(v)
	doc.UserAgent = &ecs.UserAgent{
		Original: v,
		Name:     uap.Name,
		Version:  uap.Version,
		Device: &ecs.UserAgentDevice{
			Name: uap.Device,
		},
		OS: &ecs.OS{
			Name:    uap.OS,
			Version: uap.OSVersion,
		},
	}

	return
}
