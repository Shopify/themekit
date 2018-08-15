package release

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"

	"github.com/Shopify/themekit/src/colors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	region     = "us-east-1"
	bucketName = "shopify-themekit"
)

type s3Uploader struct {
	*s3manager.Uploader
}

type uploader interface {
	File(string, io.ReadSeeker) (string, error)
	JSON(string, interface{}) error
}

func newS3Uploader(key, secret string) (*s3Uploader, error) {
	creds := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	return &s3Uploader{s3manager.NewUploader(session.New(cfg))}, nil
}

func (uploader *s3Uploader) File(fileName string, body io.ReadSeeker) (string, error) {
	colors.ColorStdOut.Printf("Uploading %s", colors.Green(fileName))
	resp, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String("public-read"),
		Key:    aws.String(fileName),
		Body:   body,
	})
	if err != nil {
		return "", err
	}

	fileURL, err := url.QueryUnescape(resp.Location)
	if err != nil {
		return "", err
	}

	colors.ColorStdOut.Printf("Complete %s", colors.Green(fileName))
	return fileURL, nil
}

func (uploader *s3Uploader) JSON(filename string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = uploader.File(filename, bytes.NewReader(jsonData))
	return err
}
