package kit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	goodEnvirontmentPath   = "../fixtures/project/valid_config.yml"
	badEnvirontmentPath    = "../fixtures/project/invalid_config.yml"
	outputEnvirontmentPath = "../fixtures/project/output.yml"
)

type EnvironmentsTestSuite struct {
	suite.Suite
	environments Environments
	errors       error
}

func (suite *EnvironmentsTestSuite) SetupTest() {
	suite.environments, suite.errors = LoadEnvironments(goodEnvirontmentPath)
}

func (suite *EnvironmentsTestSuite) TearDownTest() {
	os.Remove(outputEnvirontmentPath)
}

func (suite *EnvironmentsTestSuite) TestLoadEnvironments() {
	assert.NoError(suite.T(), suite.errors, "An error should not have been raised")
	assert.Equal(suite.T(), 3, len(suite.environments))

	_, err := LoadEnvironments(badEnvirontmentPath)
	assert.NotNil(suite.T(), err)

	_, err = LoadEnvironments(clean("./not/there.yml"))
	assert.NotNil(suite.T(), err)
}

func (suite *EnvironmentsTestSuite) TestSearchConfigPath() {
	_, ext, err := searchConfigPath(goodEnvirontmentPath)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "yml", ext)

	_, ext, err = searchConfigPath(clean("../fixtures/project/config.json"))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "json", ext)

	_, ext, err = searchConfigPath(clean("./not/there.yml"))
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), os.ErrNotExist, err)
}

func (suite *EnvironmentsTestSuite) TestSetConfiguration() {
	newConfig, _ := NewConfiguration()
	suite.environments.SetConfiguration("test", newConfig)
	assert.Equal(suite.T(), newConfig, suite.environments["test"])
}

func (suite *EnvironmentsTestSuite) TestGetConfiguration() {
	_, err := suite.environments.GetConfiguration("development")
	assert.Nil(suite.T(), err)
	_, err = suite.environments.GetConfiguration("nope")
	assert.NotNil(suite.T(), err)
}

func (suite *EnvironmentsTestSuite) TestSave() {
	err := suite.environments.Save(outputEnvirontmentPath)
	assert.Nil(suite.T(), err)
	_, err = os.Stat(outputEnvirontmentPath)
	assert.Nil(suite.T(), err)
	err = suite.environments.Save(clean("./no/where/path"))
	assert.NotNil(suite.T(), err)
}

func TestEnvironmentsTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentsTestSuite))
}
