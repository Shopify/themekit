package kittest

import (
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

var invalidConfig = template.Must(template.New("invalidConfig").Parse(`development:
  store: {{ .Domain }}
  directory: {{ .Directory }}
  access_token: abracadabra
  ignore_files:
    - hello
    - *.jpg
    - *.png
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
	f, err := os.OpenFile("config.yml", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	data := struct{ Domain, Directory string }{domain, FixtureProjectPath}
	if valid {
		if err = validConfig.Execute(f, data); err != nil {
			return err
		}
	} else {
		if err = invalidConfig.Execute(f, data); err != nil {
			return err
		}
	}
	return nil
}

// GenerateJSONConfig will generate and write a valid json config for testing
func GenerateJSONConfig(domain string) error {
	Setup()
	f, err := os.OpenFile("config.json", os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	data := struct{ Domain, Directory string }{domain, FixtureProjectPath}
	if err = jsonConfig.Execute(f, data); err != nil {
		return err
	}
	return nil
}
