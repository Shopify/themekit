package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewNotifyAdapter(t *testing.T) {
	adapter := newNotifyAdapter("")
	_, ok := adapter.(*noopNotify)
	assert.True(t, ok)

	adapter = newNotifyAdapter("note.txt")
	_, ok = adapter.(*fileNotify)
	assert.True(t, ok)

	adapter = newNotifyAdapter("http://localhost:3000/notify")
	_, ok = adapter.(*urlNotify)
	assert.True(t, ok)
}

func TestNoopAdapter(t *testing.T) {
	adapter := newNotifyAdapter("")
	ctx, _, _, _, _ := createTestCtx()
	adapter.notify(ctx, "")
}

func TestNotifyFile(t *testing.T) {
	notifyPath := filepath.Join("_testdata", "notify_file")
	adapter := newNotifyAdapter(notifyPath)
	ctx, _, _, _, _ := createTestCtx()

	os.Remove(notifyPath)
	_, err := os.Stat(notifyPath)
	assert.True(t, os.IsNotExist(err))

	adapter.notify(ctx, "")
	_, err = os.Stat(notifyPath)
	assert.Nil(t, err)
	// need to make the time different larger than milliseconds because windows
	// trucates the time and it will fail
	os.Chtimes(notifyPath, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))
	info1, err := os.Stat(notifyPath)
	assert.Nil(t, err)

	adapter.notify(ctx, "")
	info2, err := os.Stat(notifyPath)
	assert.Nil(t, err)
	assert.NotEqual(t, info1.ModTime(), info2.ModTime())
}

func TestNotifyURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		data := map[string]interface{}{}
		assert.Nil(t, json.Unmarshal(reqBody, &data))
		assert.NotNil(t, data["files"])
		files, _ := data["files"].([]interface{})
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "assets/app.js", files[0])
	}))
	defer server.Close()

	ctx, _, _, _, _ := createTestCtx()

	adapter := newNotifyAdapter(server.URL)
	adapter.notify(ctx, "assets/app.js")
}
