package main

import (
	"fmt"
	"io"
	"time"

	"github.com/telkomindonesia/httpmsg-enricher/ecs"
	ecsx "github.com/telkomindonesia/httpmsg-enricher/ecs/custom"
	"go.uber.org/multierr"
)

type enrichment struct {
	ercr *enricher

	msg     *httpRecordedMessage
	reqBody *truncatedBuffer
	resBody *truncatedBuffer

	secs []subEnricher
}

func (etx *enrichment) processRequest() (err error) {
	req, err := etx.msg.Request()
	if err != nil {
		return
	}
	defer req.Body.Close()

	etx.reqBody = newTruncatedBuffer(8 * 1024)
	w := []io.WriteCloser{etx.reqBody}
	for _, sec := range etx.secs {
		w = append(w, sec.requestBodyWriter())
	}
	if err := MultiCopy(etx.msg.req.Body, w...); err != nil {
		return err
	}
	for _, sec := range etx.secs {
		if errt := sec.processRequest(req); err != nil {
			err = multierr.Append(err, errt)
		}
	}

	return
}

func (etx *enrichment) processResponse() (err error) {
	res, err := etx.msg.Response()
	if err != nil {
		return
	}
	defer res.Body.Close()

	etx.resBody = newTruncatedBuffer(8 * 1024)
	w := []io.WriteCloser{etx.resBody}
	for _, sec := range etx.secs {
		w = append(w, sec.responseBodyWriter())
	}
	if err := MultiCopy(etx.msg.res.Body, w...); err != nil {
		return err
	}
	for _, sec := range etx.secs {
		if errt := sec.processResponse(res); err != nil {
			err = multierr.Append(err, errt)
		}
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

	doc = &ecsx.Document{
		Document: ecs.Document{
			Base: ecs.Base{
				Message: "recorded HTTP message",
			},
			ECS: ecs.ECS{
				Version: "8.3.0",
			},
			Event: &ecs.Event{
				Kind:     "event",
				Type:     []string{"access"},
				Category: []string{"web", "authentication", "network"},
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
					Method:   req.Method,
					Referrer: req.Referer(),
					HTTPMessage: ecs.HTTPMessage{
						Body: &ecs.HTTPMessageBody{
							Bytes:   int64(etx.reqBody.Len()),
							Content: etx.reqBody.String(),
						},
					},
				},
				Headers: MapStringsKeyToLower(req.Header),
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
				Headers: MapStringsKeyToLower(res.Header),
			},
		},
	}

	// ctx might be null
	{
		id := req.Header.Get("X-Request-Id")
		if ctx != nil {
			id = ctx.ID
		}
		doc.HTTP.Request.ID = id
		doc.Event.Id = id

		if ctx != nil && ctx.Durations != nil {
			doc.Document.Timestamp = ctx.Durations.Total.Start
			doc.Event.Created = &ctx.Durations.Total.Start
			doc.Event.End = ctx.Durations.Total.End
		} else {
			now := time.Now()
			doc.Document.Timestamp = now
			doc.Event.Created = &now
			doc.Event.End = &now
		}

		if ctx != nil && ctx.User != nil {
			doc.User = &ecs.User{
				Name: ctx.User.Username,
			}
		}
	}

	for _, sec := range etx.secs {
		if errt := sec.enrich(doc, etx.msg); errt != nil {
			err = multierr.Append(err, errt)
		}
	}

	return
}

func (etx enrichment) Close() (err error) {
	if errt := etx.msg.Close(); errt != nil {
		err = multierr.Append(err, errt)
	}
	for _, sec := range etx.secs {
		if errt := sec.Close(); errt != nil {
			err = multierr.Append(err, errt)
		}
	}
	return
}
