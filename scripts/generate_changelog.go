package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
	"time"
)

type Release struct {
	URL    string    `json:"html_url"`
	Tag    string    `json:"tag_name"`
	Body   string    `json:"body"`
	Date   time.Time `json:"published_at"`
	Author struct {
		Name string `json:"login"`
		URL  string `json:"url"`
	} `json:"author"`
}

const releasesURL = "https://api.github.com/repos/Shopify/themekit/releases"
const changelogTemplate = `# Change Log
All released changes to this project will be documented in this file.

{{range .AllReleases}}
## [{{.Tag}}] {{.Date.Format "Jan 02, 2006"}}
{{.URL}}

{{.Body}}
{{end}}
`

func main() {
	resp, err := http.Get(releasesURL)
	must(err)
	defer resp.Body.Close()
	jsonData, err := ioutil.ReadAll(resp.Body)
	must(err)

	releases := []Release{}
	must(json.Unmarshal(jsonData, &releases))

	tmpl, err := template.New("changelog").Parse(changelogTemplate)
	must(tmpl.Execute(os.Stdout, struct {
		AllReleases []Release
	}{releases}))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
