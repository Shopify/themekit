package kittest

import (
	"io/ioutil"
	"os"
	"text/template"
)

var validConfig = template.Must(template.New("validConfig").Parse(`default:
  password: foo
  theme_id: "2"
  store: {{ .Domain }}
  directory: {{ .Directory }}
  ignore_files:
  - charmander
  - bulbasaur
  - squirtle
development:
  password: abracadabra
  theme_id: "1"
  store: {{ .Domain }}
  directory: {{ .Directory }}
  ignore_files:
  - charmander
  - bulbasaur
  - squirtle
production:
  password: abracadabra
  theme_id: "3"
  store: {{ .Domain }}
  directory: {{ .Directory }}
  ignore_files:
  - charmander
  - bulbasaur
  - squirtle
`))

const (
	badOtherEnvConfig = `development:
  password: abracadabra
  theme_id: "1"
  store: store.myshopify.com
production:
  password: abracadabra
  store: store.myshopify.com
`

	invalidConfig = `development:
  store: store.myshopify.com
  access_token: abracadabra
  ignore_files:
    - hello
    - *.jpg
    - *.png
`
)

var proxyConfig = template.Must(template.New("proxyConfig").Parse(`development:
  password: foo
  theme_id: "2"
  store: "{{ .Domain }}"
  proxy: "http://localhost:3000"
  directory: "{{ .Directory }}"
`))

var jsonConfig = template.Must(template.New("jsonConfig").Parse(`{
  "default": {
    "password": "foo",
    "theme_id": "2",
		"directory": "{{ .Directory }}",
    "store": "{{ .Domain }}"
  },
  "development": {
    "password": "foo",
    "theme_id": "2",
		"directory": "{{ .Directory }}",
    "store": "{{ .Domain }}"
  }
}`))

// GenerateConfig will generate a config using the passed domain. To test invalid
// config pass false to valid.
func GenerateConfig(domain string, valid bool) error {
	Setup()
	if valid {
		f, err := os.OpenFile("config.yml", os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return err
		}
		data := struct{ Domain, Directory string }{domain, FixtureProjectPath}
		if err = validConfig.Execute(f, data); err != nil {
			return err
		}
	} else {
		return ioutil.WriteFile("config.yml", []byte(invalidConfig), 0777)
	}
	return nil
}

// GenerateProxyConfig will generate a config using the passed domain with proxy config.
func GenerateProxyConfig(domain string) error {
	Setup()
	f, err := os.OpenFile("config.yml", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	data := struct{ Domain, Directory string }{domain, FixtureProjectPath}
	return proxyConfig.Execute(f, data)
}

// GenerateBadMultiConfig will generate a config using the passed domain with proxy config,
// but with one invalid environment
func GenerateBadMultiConfig(domain string) error {
	return ioutil.WriteFile("config.yml", []byte(badOtherEnvConfig), 0777)
}

// GenerateJSONConfig will generate and write a valid json config for testing
func GenerateJSONConfig(domain string) error {
	Setup()
	f, err := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	data := struct{ Domain, Directory string }{domain, FixtureProjectPath}
	return jsonConfig.Execute(f, data)
}
