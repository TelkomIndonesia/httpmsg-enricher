package ecs

type GeoPoint struct {
	Lon float32 `json:"lon,omitempty"`
	Lat float32 `json:"lat,omitempty"`
}
type Geo struct {
	CityName       string    `json:"city_name,omitemtpty"`
	ContinentCode  string    `json:"continent_code,omitemtpty"`
	ContinentName  string    `json:"continent_name,omitemtpty"`
	CountryISOCode string    `json:"country_iso_code,omitemtpty"`
	CountryName    string    `json:"country_name,omitemtpty"`
	Location       *GeoPoint `json:"location,omitemtpty"`
	Name           string    `json:"name,omitemtpty"`
	PostalCode     string    `json:"postal_code,omitemtpty"`
	RegionISOCode  string    `json:"region_iso_code,omitemtpty"`
	RegionName     string    `json:"region_name,omitemtpty"`
	Timezone       string    `json:"timezone,omitemtpty"`
}
