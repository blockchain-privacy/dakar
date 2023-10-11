package clustering

import (
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestInitLogger(t *testing.T) {
	require.NotPanics(t, func() {
		InitLogger()
	})
}

func TestMain(m *testing.M) {
	InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB, testhelper.ContainerNameDB)
}
