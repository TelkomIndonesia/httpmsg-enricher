package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

func main() {
	waf, err := newCorazaWaf()
	if err != nil {
		log.Fatalf("error initializing waf: %v", err)
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("etchpass", "2NLqyX5f=-Io=oiVw0D-", ""),
		Endpoint:         aws.String("https://minio.etchpass.dev"),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
	}
	s3Client := s3.New(session.New(s3Config))

	r := gin.Default()
	r.UseRawPath = true
	r.UnescapePathValues = true
	r.GET("/records/s3/:object_key", func(c *gin.Context) {
		getInput := &s3.GetObjectInput{
			Bucket: aws.String("logs"),
			Key:    aws.String(c.Param("object_key")),
		}
		resp, err := s3Client.GetObject(getInput)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		defer resp.Body.Close()

		score, err := waf.ProcessRecord(resp.Body)
		if err != nil {
			c.String(500, err.Error())
			return
		}

		c.JSON(http.StatusOK, score)
	})

	r.GET("/records/files/:filename", func(c *gin.Context) {
		f, err := os.Open(c.Param("filename"))
		if err != nil {
			c.String(500, err.Error())
			return
		}

		score, err := waf.ProcessRecord(f)
		if err != nil {
			c.String(500, err.Error())
			return
		}

		c.JSON(http.StatusOK, score)
	})

	r.Run()
}
