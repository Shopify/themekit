package cmd

import (
	"bytes"
	"text/template"
)

type themeDiff struct {
	Created []string
	Updated []string
	Removed []string
}

var themeDiffErrorTmplt = template.Must(template.New("themeDiffError").Parse(`Unexpected changes made on remote.
Diff:
{{- if .Created }}
	New Files:
		{{- range .Created }}
		- {{ print . }}
		{{- end }}
{{- end }}
{{- if .Updated }}
	Updated Files:
		{{- range .Updated }}
		- {{ print . }}
		{{- end }}
{{- end }}
{{- if .Removed }}
	Removed Files:
		{{- range .Removed }}
		- {{ print . }}
		{{- end }}
{{- end }}

You can solve this by running 'theme download' to get the most recent copy of these files.
Running 'theme download' will overwrite any changes you have made so make your work is
commit to your VCS before doing so.

If you are certain that you want to overwrite any changes then use the --force flag
`))

func newDiff() *themeDiff {
	return &themeDiff{
		Created: []string{},
		Updated: []string{},
		Removed: []string{},
	}
}

func (diff *themeDiff) Any(destructive bool) bool {
	if !destructive {
		return len(diff.Updated) > 0
	}
	return len(diff.Created) > 0 || len(diff.Updated) > 0 || len(diff.Removed) > 0
}

func (diff *themeDiff) Error() string {
	var tpl bytes.Buffer
	themeDiffErrorTmplt.Execute(&tpl, diff)
	return tpl.String()
}
