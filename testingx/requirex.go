package testingx

import (
	"testing"

	"github.com/stretchr/testify/require"

	"log"
)

func NilAndLog(t *testing.T, err error) {
	if err != nil {
		log.Errorf("Err:%+v", err.Error())
	}

	require.Nil(t, err)
}
func NotNilAndLog(t *testing.T, err error) {
	if err != nil {
		log.Errorf("Err:%+v", err.Error())
	}

	require.NotNil(t, err)
}
