package kit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	goodEnvirontmentPath   = "../fixtures/environments/valid_config.yml"
	badEnvirontmentPath    = "../fixtures/environments/bad_pattern_config.yml"
	outputEnvirontmentPath = "../fixtures/environments/output.yml"
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

	_, err = LoadEnvironments("./not/there.yml")
	assert.NotNil(suite.T(), err)
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
	err = suite.environments.Save("./no/where/path")
	assert.NotNil(suite.T(), err)
}

func TestEnvironmentsTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentsTestSuite))
}
