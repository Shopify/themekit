package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
)

type notifyAdapter interface {
	notify(*cmdutil.Ctx, string)
}

func newNotifyAdapter(notifyPath string) notifyAdapter {
	if notifyPath == "" {
		return &noopNotify{}
	} else if u, err := url.Parse(notifyPath); err == nil && u.Scheme != "" && u.Host != "" {
		return &urlNotify{
			url: notifyPath,
			client: http.Client{
				Timeout: time.Second,
			},
		}
	}
	return &fileNotify{path: notifyPath}
}

type noopNotify struct{}

func (noop *noopNotify) notify(*cmdutil.Ctx, string) {}

type urlNotify struct {
	url    string
	client http.Client
}

func (urlNote *urlNotify) notify(ctx *cmdutil.Ctx, path string) {
	body, _ := json.Marshal(map[string]interface{}{"files": []string{path}})
	resp, err := urlNote.client.Post(urlNote.url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		ctx.Log.Printf(
			`[%s] Error while notifying webhook "%s": %s`,
			colors.Green(ctx.Env.Name),
			colors.Blue(ctx.Env.Notify),
			err,
		)
	} else {
		resp.Body.Close()
	}
}

type fileNotify struct {
	path string
}

func (fileNote *fileNotify) notify(*cmdutil.Ctx, string) {
	os.Create(fileNote.path)
	os.Chtimes(fileNote.path, time.Now(), time.Now())
}
