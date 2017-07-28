package main

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

func newS3Uploader() (*s3Uploader, error) {
	stdOut.Println(green("Connecting to S3"))
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

func (uploader *s3Uploader) file(fileName string, body io.ReadSeeker) (string, error) {
	stdOut.Println("uploading", green(fileName))
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
	stdOut.Println("uploaded", green(fileName), "successfully")
	return fileURL, nil
}

func (uploader *s3Uploader) json(filename string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = uploader.file(filename, bytes.NewReader(jsonData))
	return err
}
