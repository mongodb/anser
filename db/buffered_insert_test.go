package db

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BufferedInsertSuite struct {
	suite.Suite
}

func TestBufferedInsertSuite(t *testing.T) {
	suite.Run(t, new(BufferedInsertSuite))
}

func (s *BufferedInsertSuite) SetupTest() {

}

func (s *BufferedInsertSuite) Test() {
}
