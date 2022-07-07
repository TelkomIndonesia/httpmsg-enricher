package main

import (
	"fmt"
	"io"

	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
	"github.com/oschwald/geoip2-golang"
)

type enricher struct {
	waf   *coraza.Waf
	geoDB *geoip2.Reader
}

func newEnricher(opts ...enricherFunc) (ercr *enricher, err error) {
	ercr = &enricher{}
	for _, opt := range opts {
		if err = opt(ercr); err != nil {
			return nil, err
		}
	}
	if ercr.waf == nil {
		return nil, fmt.Errorf("No CRS are provided")
	}
	return
}

type enricherFunc func(*enricher) error

func enricherFuncNOOP(*enricher) error { return nil }

func enricherWithOptionalGeoIP(dbpath string) enricherFunc {
	db, err := geoip2.Open(dbpath)
	if err != nil {
		return enricherFuncNOOP
	}
	return func(ercr *enricher) error {
		ercr.geoDB = db
		return nil
	}
}

func enricherWithCRS(rules ...string) enricherFunc {
	return func(ercr *enricher) error {
		waf := coraza.NewWaf()
		parser, _ := seclang.NewParser(waf)
		for _, f := range rules {
			if err := parser.FromFile(f); err != nil {
				return fmt.Errorf("error reading rules file from %s: %w", f, err)
			}
		}

		ercr.waf = waf
		return nil
	}
}

func (ercr *enricher) newEnrichment(record io.Reader) *enrichment {
	return &enrichment{
		ercr: ercr,
		tx:   ercr.waf.NewTransaction(),
		msg:  newHTTPRecordedMessage(record),
	}
}

func (ercr *enricher) EnrichRecord(record io.Reader) (erc *enrichment, err error) {
	erc = ercr.newEnrichment(record)

	if err = erc.processRequest(); err != nil {
		return nil, fmt.Errorf("error parsing request: %w", err)
	}

	if err = erc.processResponse(); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return
}

func (ercr *enricher) Close() error {
	return ercr.geoDB.Close()
}
