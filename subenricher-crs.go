package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v2"
	"github.com/telkomindonesia/httpmsg-enrichment/ecs"
	ecsx "github.com/telkomindonesia/httpmsg-enrichment/ecs/custom"
)

var _ subEnricher = &crsSubEnricher{}

type crsSubEnricher struct {
	tx *coraza.Transaction
}

func (erc *crsSubEnricher) Close() error {
	erc.tx.ProcessLogging()
	return erc.tx.Clean()
}

func (erc *crsSubEnricher) requestBodyWriter() io.WriteCloser {
	return nopCloser{erc.tx.RequestBodyBuffer}
}

func (erc *crsSubEnricher) processRequest(req *http.Request) (err error) {
	tx := erc.tx

	client, port := "", 0
	spl := strings.Split(req.RemoteAddr, ":")
	if len(spl) > 0 {
		client = strings.Join(spl[0:len(spl)-1], "")
	}
	if len(spl) > 1 {
		port, _ = strconv.Atoi(spl[len(spl)-1])
	}
	tx.ProcessConnection(client, port, "", 0)

	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)

	for k, vr := range req.Header {
		for _, v := range vr {
			tx.AddRequestHeader(k, v)
		}
	}
	if req.Host != "" {
		tx.AddRequestHeader("Host", req.Host)
	}
	if len(req.TransferEncoding) > 0 {
		tx.AddRequestHeader("Transfer-Encoding", strings.Join(req.TransferEncoding, ","))
	}
	tx.ProcessRequestHeaders()

	if _, err := tx.ProcessRequestBody(); err != nil {
		return fmt.Errorf("error processing request: %w", err)
	}

	return
}

func (erc *crsSubEnricher) responseBodyWriter() io.WriteCloser {
	return nopCloser{erc.tx.ResponseBodyBuffer}
}

func (erc *crsSubEnricher) processResponse(res *http.Response) (err error) {
	tx := erc.tx

	for k, v := range res.Header {
		tx.AddResponseHeader(k, strings.Join(v, ","))
	}
	if len(res.TransferEncoding) > 0 {
		tx.AddRequestHeader("Transfer-Encoding", strings.Join(res.TransferEncoding, ","))
	}
	tx.ProcessResponseHeaders(res.StatusCode, res.Proto)

	if _, err := tx.ProcessResponseBody(); err != nil {
		return fmt.Errorf("error processing response body: %w", err)
	}

	return
}

func (erc *crsSubEnricher) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
	doc.CRS = &ecsx.CRS{
		Scores: *ecsx.NewScores(erc.tx),
	}

	if doc.Threat == nil {
		doc.Threat = &ecs.Threat{}
	}
	for _, rule := range erc.tx.MatchedRules {
		idc := ecs.ThreatIndicator{
			Description: rule.ErrorLog(0),
			IP:          net.ParseIP(rule.ClientIPAddress),
			Provider:    rule.Rule.Version,
			Type:        "network-traffic",
		}
		match := &ecs.ThreatEnrichmentMatch{
			Type:   "indicator_match_rule",
			Atomic: Truncate(rule.MatchedData.Value, 200),
		}
		atk := false
		for _, tag := range rule.Rule.Tags {
			atk = atk || strings.HasPrefix(tag, "attack-")

			pl := strings.TrimPrefix(tag, "paranoia-level/")
			if pl == "" {
				continue
			}

			i, _ := strconv.ParseInt(pl, 10, 8)
			switch i {
			case 1:
				idc.Confidence = "High"
			case 2:
				idc.Confidence = "Medium"
			case 3, 4:
				idc.Confidence = "Low"
			default:
				idc.Confidence = "Not Specified"
			}
		}
		if !atk {
			continue
		}

		doc.Threat.Enrichments = append(doc.Threat.Enrichments, ecs.ThreatEnrichments{
			Indicator: idc,
			Match:     match,
		})
	}

	return nil
}
