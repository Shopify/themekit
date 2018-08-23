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
	region = "us-east-1"
)

var (
	bucket = aws.String("shopify-themekit")
	acl    = aws.String("public-read")
)

type (
	s3Uploader struct {
		upload uploadFunc
	}
	uploadFunc func(name string, body io.Reader) (string, error)
	uploader   interface {
		File(string, io.ReadSeeker) (string, error)
		JSON(string, interface{}) error
	}
)

func newS3Uploader(key, secret string) *s3Uploader {
	creds := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	return &s3Uploader{
		upload: func(filename string, body io.Reader) (string, error) {
			u := s3manager.NewUploader(session.New(cfg))
			resp, err := u.Upload(&s3manager.UploadInput{Bucket: bucket, ACL: acl, Key: aws.String(filename), Body: body})
			if err != nil {
				return "", err
			}
			return resp.Location, nil
		},
	}
}

func (uploader *s3Uploader) File(fileName string, body io.ReadSeeker) (string, error) {
	colors.ColorStdOut.Printf("Uploading %s", colors.Green(fileName))
	location, err := uploader.upload(fileName, body)
	if err != nil {
		return "", err
	}

	fileURL, err := url.QueryUnescape(location)
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
