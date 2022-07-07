package main

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type config struct {
	S3    configS3    `envPrefix:"S3_"`
	GeoIP configGeoIP `envPrefix:"GEOIP_"`
}
type configS3 struct {
	Endpoint       string             `env:"ENDPOINT"`
	Region         string             `env:"REGION"`
	ForcePathStyle bool               `env:"FORCE_PATH_STYLE"`
	Credentials    configS3Credential `envPrefix:"CREDENTIAL_"`
	Bucket         string             `env:"BUCKET"`
}
type configS3Credential struct {
	KeyID     string `env:"ACCESS_KEY_ID"`
	SecretKey string `env:"SECRET_ACCESS_KEY"`
}
type configGeoIP struct {
	CityDBPath string `env:"CITY_DB_PATH"`
}

func newConfig() (*config, error) {
	cfg := config{}
	err := env.Parse(&cfg, env.Options{
		Prefix: "CRS_OFFLINE_",
	})
	if err != nil {
		return nil, fmt.Errorf("Error parsing config: %w", err)
	}
	return &cfg, nil
}
