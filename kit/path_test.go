package kit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PathTestSuite struct {
	suite.Suite
}

func (suite *PathTestSuite) TestPathToProject() {
	root := filepath.Join("long", "path", "to")
	tests := map[string]string{
		filepath.Join(root, "config.yml"):                            "",
		filepath.Join(root, "assets", "logo.png"):                    "assets/logo.png",
		filepath.Join(root, "node_modules", "assets", "logo.png"):    "",
		filepath.Join(root, "templates", "customers", "test.liquid"): "templates/customers/test.liquid",
		filepath.Join(root, "config", "test.liquid"):                 "config/test.liquid",
		filepath.Join(root, "layout", "test.liquid"):                 "layout/test.liquid",
		filepath.Join(root, "snippets", "test.liquid"):               "snippets/test.liquid",
		filepath.Join(root, "templates", "test.liquid"):              "templates/test.liquid",
		filepath.Join(root, "locales", "test.liquid"):                "locales/test.liquid",
		filepath.Join(root, "sections", "test.liquid"):               "sections/test.liquid",
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, pathToProject(root, input))
	}
}

func (suite *PathTestSuite) TestDirInProject() {
	root := filepath.Join("long", "path", "to")
	tests := map[string]bool{
		"": false,
		filepath.Join(root, "misc"):                false,
		filepath.Join(root, "assets"):              true,
		filepath.Join(root, "templates/customers"): true,
		filepath.Join(root, "config"):              true,
		filepath.Join(root, "layout"):              true,
		filepath.Join(root, "snippets"):            true,
		filepath.Join(root, "templates"):           true,
		filepath.Join(root, "locales"):             true,
		filepath.Join(root, "sections"):            true,
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, isProjectDirectory(root, input), input)
	}
}

func (suite *PathTestSuite) TestPathInProject() {
	root := filepath.Join("long", "path", "to")
	tests := map[string]bool{
		"": false,
		filepath.Join(root, "config.yml"):                            false,
		filepath.Join(root, "misc", "other.html"):                    false,
		filepath.Join(root, "assets", "logo.png"):                    true,
		filepath.Join(root, "node_modules", "assets", "logo.png"):    false,
		filepath.Join(root, "templates", "customers", "test.liquid"): true,
		filepath.Join(root, "config", "test.liquid"):                 true,
		filepath.Join(root, "layout", "test.liquid"):                 true,
		filepath.Join(root, "snippets", "test.liquid"):               true,
		filepath.Join(root, "templates", "test.liquid"):              true,
		filepath.Join(root, "locales", "test.liquid"):                true,
		filepath.Join(root, "sections", "test.liquid"):               true,
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, pathInProject(root, input), input)
	}
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}
