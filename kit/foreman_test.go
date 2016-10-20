package kit

import (
	"testing"

	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ForemanTestSuite struct {
	suite.Suite
}

func (suite *ForemanTestSuite) TestNewForeman() {
}

func (suite *ForemanTestSuite) TestRestart() {
}

func (suite *ForemanTestSuite) TestIssueWork() {
}

func (suite *ForemanTestSuite) TestHalt() {
}

func TestForemanTestSuite(t *testing.T) {
	suite.Run(t, new(ForemanTestSuite))
}
