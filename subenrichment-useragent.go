package main

import (
	"fmt"
	"net/http"

	ua "github.com/mileusna/useragent"
	"github.com/telkomindonesia/crs-offline/ecs"
	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

var _ subEnrichment = &uaEnrichment{}

type uaEnrichment struct{}

func (uaEnrichment) requestBodyWriter() closableWriter              { return wnoop }
func (uaEnrichment) processRequest(req *http.Request) (err error)   { return }
func (uaEnrichment) responseBodyWriter() closableWriter             { return wnoop }
func (uaEnrichment) processResponse(res *http.Response) (err error) { return }
func (uaEnrichment) Close() (err error)                             { return }

func (uaEnrichment) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
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
