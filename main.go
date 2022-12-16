package main

import (
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3Config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func newS3Client(cfg *config) (*s3.Client, error) {
	s3Cfg, err := s3Config.LoadDefaultConfig(context.TODO(),
		s3Config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: cfg.S3.Endpoint}, nil
				})),
		s3Config.WithRegion(cfg.S3.Region),
		s3Config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.S3.Credentials.KeyID, cfg.S3.Credentials.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(s3Cfg), err
}

func main() {
	cfg, err := newConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	ercr, err := newEnricher(
		enricherWithCRS("crs/coraza.conf", "crs/crs-setup.conf", "crs/rules/*.conf"),
		enricherWithOptionalGeoIP(cfg.GeoIP.CityDBPath),
	)
	if err != nil {
		log.Fatalf("error initializing enricher: %v", err)
	}

	s3Client, err := newS3Client(cfg)
	if err != nil {
		log.Fatalf("error initializing s3 client: %v", err)
	}

	r := gin.Default()
	r.UseRawPath = true
	r.UnescapePathValues = true
	r.GET("/ecs/s3/:object_key", func(c *gin.Context) {
		getInput := &s3.GetObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(c.Param("object_key")),
		}
		resp, err := s3Client.GetObject(context.TODO(), getInput)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		defer resp.Body.Close()

		body := resp.Body
		if strings.Contains(*resp.ContentType, "application/gzip") {
			body, err = gzip.NewReader(body)
			if err != nil {
				c.String(500, err.Error())
				return
			}
			defer body.Close()
		}

		erc, err := ercr.EnrichRecord(body)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		defer erc.Close()

		ecs, err := erc.toECS()
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(http.StatusOK, ecs)
	})

	r.GET("/ecs/files/:filename", func(c *gin.Context) {
		f, err := os.Open("./" + path.Clean(c.Param("filename")))
		if err != nil {
			c.String(500, err.Error())
			return
		}
		defer f.Close()

		body := io.ReadCloser(f)
		if strings.HasSuffix(c.Param("filename"), ".gz") {
			body, err = gzip.NewReader(body)
			if err != nil {
				c.String(500, err.Error())
				return
			}
			defer body.Close()
		}

		er, err := ercr.EnrichRecord(f)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		defer er.Close()

		ecs, err := er.toECS()
		if err != nil {
			c.String(500, err.Error())
			return
		}
		c.JSON(http.StatusOK, ecs)
	})

	r.Run()
}
