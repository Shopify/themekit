// Package static allows for the bundling of static assets into a binary. This
// allows themekit to bundle files from a theme for generating a new theme.
package static

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

const tmpl = `package {{.PackageName}}

import "github.com/Shopify/themekit/src/static"

func init() {
	static.Register("{{.Data}}")
}`

// Bundle takes a directory path and will compress all the data into a single
// go file with zip data at the dst path.
func Bundle(src, dst string) error {
	buffer, err := compressData(src)
	if err != nil {
		return err
	}
	return writeOutTemplate(dst, buffer)
}

func compressData(src string) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buffer)
	if err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		f, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		return err
	}); err != nil {
		return nil, err
	}
	err := zipWriter.Close()
	return sanitizeData(buffer), err
}

func sanitizeData(zipData *bytes.Buffer) *bytes.Buffer {
	dest := new(bytes.Buffer)
	for _, b := range zipData.Bytes() {
		switch {
		case b == '\n':
			dest.WriteString(`\n`)
		case b == '\\':
			dest.WriteString(`\\`)
		case b == '"':
			dest.WriteString(`\"`)
		case b >= 32 && b <= 126 || b == '\t':
			dest.WriteByte(b)
		default:
			fmt.Fprintf(dest, "\\x%02x", b)
		}
	}
	return dest
}

func writeOutTemplate(dst string, data *bytes.Buffer) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	bundleTemplate, err := template.New("bundle").Parse(tmpl)
	if err != nil {
		return err
	}

	pkg := path.Base(path.Dir(dst))
	return bundleTemplate.Execute(file, struct {
		PackageName string
		Data        *bytes.Buffer
	}{PackageName: pkg, Data: data})
}
