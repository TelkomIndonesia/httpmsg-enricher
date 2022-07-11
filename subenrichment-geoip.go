package main

import (
	"fmt"
	"net/http"

	"github.com/oschwald/geoip2-golang"
	"github.com/telkomindonesia/crs-offline/ecs"
	ecsx "github.com/telkomindonesia/crs-offline/ecs/custom"
)

var _ subEnrichment = &geoipEnrichment{}

type geoipEnrichment struct {
	db *geoip2.Reader
}

func (g *geoipEnrichment) requestBodyWriter() closableWriter              { return wnoop }
func (g *geoipEnrichment) processRequest(req *http.Request) (err error)   { return }
func (g *geoipEnrichment) responseBodyWriter() closableWriter             { return wnoop }
func (g *geoipEnrichment) processResponse(res *http.Response) (err error) { return }

func (g *geoipEnrichment) enrich(doc *ecsx.Document, msg *httpRecordedMessage) (err error) {
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

func (g *geoipEnrichment) Close() error {
	return g.db.Close()
}
