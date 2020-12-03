package lint

import (
	"io"

	"github.com/osteele/liquid/render"
)

var (
	capturedTags    = []string{"assign", "include", "layout", "section", "break", "continue", "cycle"}
	capturedFilters = []string{
		"default", "compact", "join", "map", "reverse", "sort", "first", "last", "uniq",
		"date", "abs", "ceil", "floor", "modulo", "minus", "plus", "times", "divided_by", "round",
		"size", "append", "capitalize", "downcase", "escape", "escape_once", "newline_to_br", "prepend",
		"remove", "remove_first", "replace", "replace_first", "sort_natural", "slice", "split",
		"strip_html", "strip_newlines", "strip", "lstrip", "rstrip", "truncate", "truncatewords",
		"upcase", "url_encode", "url_decode", "inspect", "type", "t", "img_url", "date",
	}
	capturedBlocks = [][]string{
		{"form"}, {"paginate"}, {"style"}, {"schema"}, {"javascript"}, {"stylesheet"},
		{"capture"}, {"comment"}, {"raw"}, {"tablerow"}, {"unless"},
		{"for", "else"}, {"case", "when", "else"}, {"if", "else", "elseif"},
	}
)

func shopifyFilters(cfg *render.Config, checks CheckCollection) {
	for _, tag := range capturedTags {
		cfg.AddTag(tag, func(source string) (func(io.Writer, render.Context) error, error) {
			return func(w io.Writer, ctx render.Context) error {
				checks.Call(Tag)
				return nil
			}, nil
		})
	}

	for _, filter := range capturedFilters {
		cfg.AddFilter(filter, func(value interface{}) interface{} {
			checks.Call(Filter)
			return value
		})
	}

	for _, blockdef := range capturedBlocks {
		block := cfg.AddBlock(blockdef[0])
		for i := 1; i < len(blockdef); i++ {
			block = block.Clause(blockdef[i])
		}
		block.Compiler(func(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
			return func(w io.Writer, ctx render.Context) error {
				return nil
			}, nil
		})
	}
}
