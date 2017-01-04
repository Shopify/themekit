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
	tests := map[string]string{
		filepath.Join("long", "path", "to", "config.yml"):                            "",
		filepath.Join("long", "path", "to", "assets", "logo.png"):                    "assets/logo.png",
		filepath.Join("long", "path", "to", "templates", "customers", "test.liquid"): "templates/customers/test.liquid",
		filepath.Join("long", "path", "to", "config", "test.liquid"):                 "config/test.liquid",
		filepath.Join("long", "path", "to", "layout", "test.liquid"):                 "layout/test.liquid",
		filepath.Join("long", "path", "to", "snippets", "test.liquid"):               "snippets/test.liquid",
		filepath.Join("long", "path", "to", "templates", "test.liquid"):              "templates/test.liquid",
		filepath.Join("long", "path", "to", "locales", "test.liquid"):                "locales/test.liquid",
		filepath.Join("long", "path", "to", "sections", "test.liquid"):               "sections/test.liquid",
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, pathToProject(input))
	}
}

func (suite *PathTestSuite) TestDirInProject() {
	tests := map[string]bool{
		"": false,
		filepath.Join("long", "path", "to", "misc"):                false,
		filepath.Join("long", "path", "to", "assets"):              true,
		filepath.Join("long", "path", "to", "templates/customers"): true,
		filepath.Join("long", "path", "to", "config"):              true,
		filepath.Join("long", "path", "to", "layout"):              true,
		filepath.Join("long", "path", "to", "snippets"):            true,
		filepath.Join("long", "path", "to", "templates"):           true,
		filepath.Join("long", "path", "to", "locales"):             true,
		filepath.Join("long", "path", "to", "sections"):            true,
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, isProjectDirectory(input), input)
	}
}

func (suite *PathTestSuite) TestPathInProject() {
	tests := map[string]bool{
		"": false,
		filepath.Join("long", "path", "to", "config.yml"):                            false,
		filepath.Join("long", "path", "to", "misc", "other.html"):                    false,
		filepath.Join("long", "path", "to", "assets", "logo.png"):                    true,
		filepath.Join("long", "path", "to", "templates", "customers", "test.liquid"): true,
		filepath.Join("long", "path", "to", "config", "test.liquid"):                 true,
		filepath.Join("long", "path", "to", "layout", "test.liquid"):                 true,
		filepath.Join("long", "path", "to", "snippets", "test.liquid"):               true,
		filepath.Join("long", "path", "to", "templates", "test.liquid"):              true,
		filepath.Join("long", "path", "to", "locales", "test.liquid"):                true,
		filepath.Join("long", "path", "to", "sections", "test.liquid"):               true,
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, pathInProject(input), input)
	}
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}
