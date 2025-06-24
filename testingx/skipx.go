package testingx

import (
	"os"
	"strings"
	"testing"
)

const (
	TagRunInPipeline = "TagRunInPipeline"
)

func Skip(t *testing.T, flag string) {
	if !strings.Contains(os.Getenv("GO_TEST_ENV"), flag) {
		t.Skip("Skip, request for " + flag)
	}
}

func SkipIf(t *testing.T, flag string) {
	if strings.Contains(os.Getenv("GO_TEST_ENV"), flag) {
		t.Skip("Skip, reason: " + flag)
	}
}
