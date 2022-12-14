package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/oschwald/geoip2-golang"
	"github.com/telkomindonesia/httpmsg-enricher/ecs"
	ecsx "github.com/telkomindonesia/httpmsg-enricher/ecs/custom"
)

var _ subEnricher = &geoipEnricher{}

type geoipEnricher struct {
	db *geoip2.Reader
}

func (g *geoipEnricher) requestBodyWriter() io.WriteCloser              { return nopwc }
func (g *geoipEnricher) processRequest(req *http.Request) (err error)   { return }
func (g *geoipEnricher) responseBodyWriter() io.WriteCloser             { return nopwc }
func (g *geoipEnricher) processResponse(res *http.Response) (err error) { return }

func (g *geoipEnricher) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
	if g.db == nil {
		return fmt.Errorf("no geodatabase instantiated")
	}

	ctx, err := msg.Context()
	if err != nil {
		return fmt.Errorf("error geting context: %w", err)
	}
	if ctx == nil || ctx.Connection == nil {
		return
	}

	record, err := g.db.City(ctx.Connection.Client.IP)
	if err != nil {
		return fmt.Errorf("error geting city information: %w", err)
	}

	doc.Client = &ecs.Endpoint{
		IP: ctx.Connection.Client.IP,
		Geo: ecs.Geo{
			CityName: record.City.Names["en"],

			CountryName:    record.Country.Names["en"],
			CountryISOCode: record.Country.IsoCode,

			ContinentName: record.Continent.Names["en"],
			ContinentCode: record.Continent.Code,
			Location: &ecs.GeoPoint{
				Lon: record.Location.Longitude,
				Lat: record.Location.Latitude,
			},
			PostalCode: record.Postal.Code,
			Timezone:   record.Location.TimeZone,
		},
	}
	return nil
}

func (g *geoipEnricher) Close() error {
	return g.db.Close()
}
