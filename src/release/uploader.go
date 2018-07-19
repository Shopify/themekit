package release

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	region     = "us-east-1"
	bucketName = "shopify-themekit"
)

type deploySecrets struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

type s3Uploader struct {
	*s3manager.Uploader
}

type uploader interface {
	File(string, io.ReadSeeker) (string, error)
	JSON(string, interface{}) error
}

func newS3Uploader() (*s3Uploader, error) {
	raw, err := ioutil.ReadFile(".env")
	if err != nil {
		return nil, err
	}

	var secrets deploySecrets
	err = json.Unmarshal(raw, &secrets)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentials(secrets.Key, secrets.Secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)

	return &s3Uploader{s3manager.NewUploader(session.New(cfg))}, nil
}

func (uploader *s3Uploader) File(fileName string, body io.ReadSeeker) (string, error) {
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
