package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

const outputPath = "../fixtures/fomatted_output.json"

type DownloadTestSuite struct {
	suite.Suite
}

func (suite *DownloadTestSuite) TearDownTest() {
	os.RemoveAll("../fixtures/output")
	os.RemoveAll("../fixtures/download")
}

func (suite *DownloadTestSuite) TestDownloadWithFileNames() {
	defer os.Remove("../fixtures/project/assets/hello.txt")
	defer os.Remove("../fixtures/project/assets/hello.txt")
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		if "fields=key,attachment,value&asset[key]=assets/hello.txt" == r.URL.RawQuery {
			fmt.Fprintf(w, jsonFixture("responses/asset"))
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "404")
		}
	})
	defer server.Close()

	err := download(client, []string{"assets/hello.txt"})
	assert.Nil(suite.T(), err)

	err = download(client, []string{"output/nope.txt"})
	assert.NotNil(suite.T(), err)
}

func (suite *DownloadTestSuite) TestDownloadAll() {
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	client.Config.Directory = "../fixtures/download"
	os.MkdirAll(client.Config.Directory, 7777)
	defer os.Remove(client.Config.Directory)

	err := download(client, []string{})
	assert.Nil(suite.T(), err)
}

func (suite *DownloadTestSuite) TestWriteToDisk() {
	asset := kit.Asset{Key: "output/blah.txt", Value: "this is content"}

	err := writeToDisk("../nope", asset)
	assert.NotNil(suite.T(), err)

	err = writeToDisk("../fixtures", asset)
	assert.Nil(suite.T(), err)
}

func (suite *DownloadTestSuite) TestGetAssetContents() {
	data, err := getAssetContents(kit.Asset{Value: "this is content"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 15, len(data))

	data, err = getAssetContents(kit.Asset{Attachment: "this is bad content"})
	assert.NotNil(suite.T(), err)

	data, err = getAssetContents(kit.Asset{Attachment: base64.StdEncoding.EncodeToString([]byte("this is bad content"))})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 19, len(data))
	assert.Equal(suite.T(), []byte("this is bad content"), data)
}

func (suite *DownloadTestSuite) TestFormatWrite() {
	file, _ := os.Create(outputPath)

	n, err := formatWrite(file, []byte{})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, n)

	n, err = formatWrite(file, []byte("{\"test\":\"one\"}"))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 19, n)
	file.Close()

	file, _ = os.Open(outputPath)
	buffer, err := ioutil.ReadAll(file)
	assert.Equal(suite.T(), `{
  "test": "one"
}`, string(buffer))

	os.Remove(outputPath)
}

func TestDownloadTestSuite(t *testing.T) {
	suite.Run(t, new(DownloadTestSuite))
}

func fileFixture(name string) *os.File {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, _ := os.Open(path)
	return file
}

func jsonFixture(name string) string {
	bytes, err := ioutil.ReadAll(fileFixture(name))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}
