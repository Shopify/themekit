package cmd

// import (
//	"os"
//	"path/filepath"
//	"testing"

//	"github.com/stretchr/testify/assert"
// )

// func TestDownloadWithFileNames(t *testing.T) {
//	server := newTestServer()
//	defer server.Close()
//	defer os.Remove(filepath.Join(fixtureProjectPath, "assets", "hello.txt"))

//	client, err := getClient()
//	if assert.Nil(t, err) {
//		err = download(client, []string{"assets/hello.txt"})
//		assert.Nil(t, err)
//	}
// }

// func TestDownloadWithReadOnly(t *testing.T) {
//	server := newTestServer()
//	defer server.Close()
//	defer os.Remove(filepath.Join(fixtureProjectPath, "assets", "hello.txt"))

//	client, err := getClient()
//	if assert.Nil(t, err) {
//		client.Config.ReadOnly = true
//		err = download(client, []string{"output/nope.txt"})
//		assert.Nil(t, err)
//	}
// }

// func TestDownloadAll(t *testing.T) {
//	server := newTestServer()
//	defer server.Close()
//	defer os.Remove(filepath.Join(fixturesPath, "download"))

//	client, err := getClient()
//	if assert.Nil(t, err) {
//		client.Config.Directory = filepath.Join(fixturesPath, "download")
//		os.MkdirAll(client.Config.Directory, 7777)
//		defer os.Remove(client.Config.Directory)
//		assert.Nil(t, download(client, []string{}))
//	}
// }
