package testhelper

import (
	"os"
	"testing"
)

const EnvCIFlag = "CI_ACTIVE"

func IsCIActive() bool {
	return os.Getenv(EnvCIFlag) != ""
}

func SkipIfNotCI(t *testing.T) {
	if !IsCIActive() {
		t.Skip("skipping CI test")
	}
}
