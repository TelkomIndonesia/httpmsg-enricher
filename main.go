package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := newConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.S3.Credentials.KeyID, cfg.S3.Credentials.SecretKey, ""),
		Endpoint:         aws.String(cfg.S3.Endpoint),
		Region:           aws.String(cfg.S3.Region),
		S3ForcePathStyle: aws.Bool(cfg.S3.ForcePathStyle),
	}
	s3Client := s3.New(session.New(s3Config))

	ercr, err := newEnricher(
		enricherWithCRS("crs/coraza.conf", "crs/crs-setup.conf", "crs/rules/*.conf"),
		enricherWithOptionalGeoIP(cfg.GeoIP.CityDBPath),
	)
	if err != nil {
		log.Fatalf("error initializing enricher: %v", err)
	}

	r := gin.Default()
	r.UseRawPath = true
	r.UnescapePathValues = true
	r.GET("/ecs/s3/:object_key", func(c *gin.Context) {
		getInput := &s3.GetObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(c.Param("object_key")),
		}
		resp, err := s3Client.GetObject(getInput)
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
