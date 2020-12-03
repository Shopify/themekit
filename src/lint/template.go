package lint

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

type Template struct {
	cfg          *render.Config
	root         string
	path         string
	Source       string
	RelativePath string
	Name         string
	Type         string
}

func newTemplate(cfg *render.Config, root, path string) *Template {
	source, _ := ioutil.ReadFile(path)
	relativePath, _ := filepath.Rel(root, path)
	template := &Template{
		cfg:          cfg,
		root:         root,
		path:         path,
		Source:       string(source),
		RelativePath: relativePath,
		Name:         strings.ReplaceAll(relativePath, ".liquid", ""),
		Type:         "other",
	}
	if strings.HasPrefix(relativePath, "sections") {
		template.Type = "section"
	} else if strings.HasPrefix(relativePath, "snippets") {
		template.Type = "snippet"
	} else if strings.HasPrefix(relativePath, "templates") {
		template.Type = "template"
	}
	return template
}

func (template *Template) Lines() []string {
	return strings.Split(template.Source, "\n")
}

func (template *Template) Excerpt(line int) string {
	return strings.TrimSpace(template.Lines()[line-1])
}

func (template *Template) Check(checks CheckCollection) error {
	checks.Call(BeginDoc)
	root, err := template.cfg.Parse(template.Source, parser.SourceLoc{Pathname: template.path})
	if err != nil {
		return err
	}
	follow(root, template.cfg)
	return nil
}

func follow(node parser.ASTNode, cfg *render.Config) {
	switch v := node.(type) {
	case *parser.ASTBlock:
		for _, node := range v.Body {
			follow(node, cfg)
		}
		for _, node := range v.Clauses {
			follow(node, cfg)
		}
	case *parser.ASTObject:
		v.Expr.Evaluate(expressions.NewContext(nil, expressions.NewConfig()))
		fmt.Println("ASTObject", v.Token)
	case *parser.ASTRaw:
		fmt.Println("ASTRaw", len(v.Slices))
	case *parser.ASTSeq:
		for _, node := range v.Children {
			follow(node, cfg)
		}
	case *parser.ASTTag:
		fmt.Println("ASTTag", v.Token)
	case *parser.ASTText:
		//fmt.Println("ASTText", v.Token)
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}
