package main

import (
	"fmt"
	"strings"

	"github.com/telkomindonesia/crs-offline/ecs"
	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

type enrichment struct {
	ercr *enricher

	msg     *httpRecordedMessage
	reqBody *truncatedBuffer
	resBody *truncatedBuffer

	secs []subEnrichment
}

func (etx *enrichment) processRequest() (err error) {
	req, err := etx.msg.Request()
	if err != nil {
		return
	}

	etx.reqBody = newTruncatedBuffer(1024 * 1024)
	w := []closableWriter{}
	for _, sec := range etx.secs {
		w = append(w, sec.requestBodyWriter())
	}
	MultiCopy(etx.msg.req.Body, w...)
	for _, sec := range etx.secs {
		sec.processRequest(req)
	}

	return
}

func (etx *enrichment) processResponse() (err error) {
	res, err := etx.msg.Response()
	if err != nil {
		return
	}

	// response body
	etx.resBody = newTruncatedBuffer(1024 * 1024)
	w := []closableWriter{}
	for _, sec := range etx.secs {
		w = append(w, sec.responseBodyWriter())
	}
	MultiCopy(etx.msg.res.Body, w...)
	for _, sec := range etx.secs {
		sec.processResponse(res)
	}

	return
}

func (etx *enrichment) toECS() (doc *ecsx.Document, err error) {
	req, res := etx.msg.req, etx.msg.res
	if req == nil || res == nil {
		return nil, fmt.Errorf("Please invoke ProcessRequest() and ProcessResponse() first.")
	}
	ctx, err := etx.msg.Context()
	if err != nil {
		return nil, fmt.Errorf("error geting context: %w", err)
	}

	toLower := func(m map[string][]string) map[string][]string {
		nm := map[string][]string{}
		for k, v := range m {
			nm[strings.ToLower(k)] = v
		}
		return nm
	}

	doc = &ecsx.Document{
		Document: ecs.Document{
			Base: ecs.Base{
				Message:   "recorded HTTP message",
				Timestamp: ctx.Durations.Proxy.Start,
			},
			ECS: ecs.ECS{
				Version: "8.3.0",
			},
			Event: &ecs.Event{
				Kind:     "event",
				Type:     []string{"access"},
				Category: []string{"web", "authentication", "network"},
				Created:  &ctx.Durations.Total.Start,
				End:      ctx.Durations.Total.End,
				Id:       ctx.ID,
			},
			URL: &ecs.URL{
				Domain:   req.Host,
				Full:     req.URL.String(),
				Original: req.URL.String(),
				Query:    req.URL.Query().Encode(),
				Fragment: req.URL.Fragment,
				Path:     req.URL.Path,
				Scheme:   req.URL.Scheme,
			},
			Threat: &ecs.Threat{
				Enrichments: []ecs.ThreatEnrichments{},
			},
		},

		HTTP: &ecsx.HTTP{
			HTTP: ecs.HTTP{
				Version: fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor),
			},
			Request: &ecsx.HTTPRequest{
				HTTPRequest: ecs.HTTPRequest{
					ID:       ctx.ID,
					Method:   req.Method,
					Referrer: req.Referer(),
					HTTPMessage: ecs.HTTPMessage{
						Body: &ecs.HTTPMessageBody{
							Bytes:   int64(etx.reqBody.Len()),
							Content: etx.reqBody.String(),
						},
					},
				},
				Headers: toLower(req.Header),
			},
			Response: &ecsx.HTTPResponse{
				HTTPResponse: ecs.HTTPResponse{
					StatusCode: res.StatusCode,
					HTTPMessage: ecs.HTTPMessage{
						Body: &ecs.HTTPMessageBody{
							Bytes:   int64(etx.resBody.Len()),
							Content: etx.resBody.String(),
						},
					},
				},
				Headers: toLower(res.Header),
			},
		},
	}

	if ctx != nil && ctx.Credential != nil {
		doc.User = &ecs.User{
			Name: ctx.Credential.Username,
		}
	}

	for _, sec := range etx.secs {
		sec.enrich(doc, etx.msg)
	}

	return
}

func (etx enrichment) Close() (err error) {
	etx.msg.Close()
	for _, sec := range etx.secs {
		sec.Close()
	}
	return
}
