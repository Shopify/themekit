package lint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/osteele/liquid/render"
)

type Theme struct {
	all       []*Template
	Templates []*Template
	Sections  []*Template
	Snippets  []*Template
}

func newTheme(cfg *render.Config, root string, checks CheckCollection) Theme {
	newTheme := Theme{}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".liquid" {
			template := newTemplate(cfg, root, path)
			newTheme.all = append(newTheme.all, template)
			if template.Type == "template" {
				newTheme.Templates = append(newTheme.Templates, template)
			} else if template.Type == "section" {
				newTheme.Sections = append(newTheme.Sections, template)
			} else if template.Type == "snippet" {
				newTheme.Snippets = append(newTheme.Snippets, template)
			}
			if err := template.Check(checks); err != nil {
				fmt.Println(err)
			}
		}
		return nil
	})
	return newTheme
}
