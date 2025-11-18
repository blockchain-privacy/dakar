// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"backend/analytics/graph"
	"backend/db"
	"backend/db/status"
	"backend/db/workspace"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewWorker(t *testing.T) {
	w := NewWorker(NewMutex(), nil, nil)
	require.NotNil(t, w)
}

func TestWorker_SetLoopInterval(_ *testing.T) {
	w := NewWorker(NewMutex(), nil, nil)
	w.SetLoopInterval(1)
	w.SetLoopInterval(time.Second * 100)
}

func TestWorker_work(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UsePrivacyFile)

	// create test data
	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	startDate1, err := time.Parse(time.RFC3339, "2021-07-22T00:00:00Z")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-07-23T00:00:00Z")
	require.NoError(t, err)

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)

	opt := workspace.TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: &val1},
		InputRange:  &workspace.AmountRange{Min: &valPoint01, Max: &valPoint1},
		OutputRange: &workspace.AmountRange{Min: &val1, Max: &valPoint1},
	}

	m := NewMutex()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// insert 3 selectors into db
	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	wrapper := graph.NewWrapper(ctx, dbHandle)
	wrapper.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, status.SetLastClassifiedBlockID(ctx, dbHandle, int64(db.ClassifierFileLastBlock)))
	require.NoError(t, wrapper.LoadGraphs(graph.NewDashConfig()))

	w := NewWorker(NewMutex(), dbHandle, wrapper)
	w.RegisterMetrics(prometheus.NewRegistry())
	w.SetLoopInterval(1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		w.Start(ctx)
	}()

	// if the number of waiting selectors is 0, then the worker has finished
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

	// stop worker
	cancel()
	wg.Wait()
}
