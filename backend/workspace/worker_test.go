package workspace

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/status"
	"backend/db/workspace"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewWorker(t *testing.T) {
	w := NewWorker(NewMutex(), nil, nil)
	require.NotNil(t, w)
	require.NotNil(t, w.activeMutex)
}

func TestWorker_Start(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wrapper := graph.NewWrapper(ctx, dbHandle)
	w := NewWorker(NewMutex(), dbHandle, wrapper)
	w.RegisterMetrics(prometheus.NewRegistry())

	require.True(t, w.Start(ctx))
	// worker already active
	require.False(t, w.Start(ctx))
}

func TestWorker_SetLoopInterval(_ *testing.T) {
	w := NewWorker(NewMutex(), nil, nil)
	w.SetLoopInterval(1)
	w.SetLoopInterval(time.Second * 100)
}

func TestWorker_Stop(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wrapper := graph.NewWrapper(ctx, dbHandle)
	w := NewWorker(NewMutex(), dbHandle, wrapper)
	w.RegisterMetrics(prometheus.NewRegistry())

	// stop without start
	w.Stop()

	require.True(t, w.Start(ctx))
	w.Stop()
	require.False(t, w.IsActive())
}

func TestWorker_work(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UsePrivacyFile)

	// create test data
	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	startDate1, err := time.Parse(time.RFC3339, "2021-07-22T00:00:00Z")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-07-23T00:00:00Z")
	require.NoError(t, err)

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)

	opt := workspace.Options{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: &val1},
		InputRange:  &workspace.AmountRange{Min: &valPoint01, Max: &valPoint1},
		OutputRange: &workspace.AmountRange{Min: &val1, Max: &valPoint1},
	}

	m := NewMutex()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// insert 3 selectors into db
	_, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTransactionProperties, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTransactionProperties, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTransactionProperties, "", workspaceUID, userUID)
	require.NoError(t, err)

	wrapper := graph.NewWrapper(ctx, dbHandle)
	wrapper.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, status.SetLastClassifiedBlockID(dbHandle, uint64(testhelper.ClassifierFileLastBlock)))
	require.NoError(t, wrapper.LoadGraphs())

	w := NewWorker(NewMutex(), dbHandle, wrapper)
	w.RegisterMetrics(prometheus.NewRegistry())
	w.SetLoopInterval(1)
	require.True(t, w.Start(ctx))

	// if the number of waiting selectors is 0, then the worker has finsihed
	now := time.Now()
	for {
		time.Sleep(time.Millisecond * 100)
		selectorsWaiting, err := workspace.GetWaitingSelectors(ctx, dbHandle, 20)
		require.NoError(t, err)

		if len(selectorsWaiting) == 0 {
			break
		}

		// prevent infinite loop in case something went wrong
		if time.Since(now) > time.Second*30 {
			t.Error("worker took to long to finish work")
			break
		}
	}
}
