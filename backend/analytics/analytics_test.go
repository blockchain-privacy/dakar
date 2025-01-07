package analytics

import (
	"backend/testhelper"
	"testing"
)

var dbHandle = &testhelper.TestDB{}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, dbHandle)
}
