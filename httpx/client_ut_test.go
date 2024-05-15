package httpx

import (
	"testing"

	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite
}

func (ts *ClientTestSuite) SetupSuite() {
}

func (ts *ClientTestSuite) TearDownSuite() {
}

func TestClientTestSuite_UT(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
