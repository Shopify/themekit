package file

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathToProject(t *testing.T) {
	root := filepath.Join("long", "path", "to")
	tests := map[string]string{
		filepath.Join(root, "config.yml"):                            "",
		filepath.Join(root, "assets", "logo.png"):                    "assets/logo.png",
		filepath.Join(root, "node_modules", "assets", "logo.png"):    "",
		filepath.Join(root, "templates", "customers", "test.liquid"): "templates/customers/test.liquid",
		filepath.Join(root, "pages", "customers", "test.liquid"):     "pages/customers/test.liquid",
		filepath.Join(root, "config", "test.liquid"):                 "config/test.liquid",
		filepath.Join(root, "layout", "test.liquid"):                 "layout/test.liquid",
		filepath.Join(root, "snippets", "test.liquid"):               "snippets/test.liquid",
		filepath.Join(root, "templates", "test.liquid"):              "templates/test.liquid",
		filepath.Join(root, "locales", "test.liquid"):                "locales/test.liquid",
		filepath.Join(root, "sections", "test.liquid"):               "sections/test.liquid",
	}
	for input, expected := range tests {
		assert.Equal(t, expected, pathToProject(root, input))
	}
}

func TestDirInProject(t *testing.T) {
	root := filepath.Join("long", "path", "to")
	tests := map[string]bool{
		"":                                         false,
		filepath.Join(root, "assets"):              true,
		filepath.Join(root, "config"):              true,
		filepath.Join(root, "content"):             true,
		filepath.Join(root, "css"):                 false,
		filepath.Join(root, "frame"):               true,
		filepath.Join(root, "layout"):              true,
		filepath.Join(root, "locales"):             true,
		filepath.Join(root, "misc"):                false,
		filepath.Join(root, "node_modules"):        false,
		filepath.Join(root, "pages"):               true,
		filepath.Join(root, "pages/customers"):     true,
		filepath.Join(root, "sections"):            true,
		filepath.Join(root, "snippets"):            true,
		filepath.Join(root, "templates"):           true,
		filepath.Join(root, "templates/customers"): true,
	}
	for input, expected := range tests {
		assert.Equal(t, expected, isProjectDirectory(root, input), input)
	}
}

func TestPathInProject(t *testing.T) {
	root := filepath.Join("long", "path", "to")
	tests := map[string]bool{
		"":                                false,
		filepath.Join(root, "config.yml"): false,
		filepath.Join(root, "misc", "other.html"):                    false,
		filepath.Join(root, "assets", "logo.png"):                    true,
		filepath.Join(root, "node_modules", "assets", "logo.png"):    false,
		filepath.Join(root, "pages", "customers", "test.liquid"):     true,
		filepath.Join(root, "templates", "customers", "test.liquid"): true,
		filepath.Join(root, "config", "test.liquid"):                 true,
		filepath.Join(root, "layout", "test.liquid"):                 true,
		filepath.Join(root, "snippets", "test.liquid"):               true,
		filepath.Join(root, "templates", "test.liquid"):              true,
		filepath.Join(root, "locales", "test.liquid"):                true,
		filepath.Join(root, "sections", "test.liquid"):               true,
	}
	for input, expected := range tests {
		assert.Equal(t, expected, pathInProject(root, input), input)
	}
}
