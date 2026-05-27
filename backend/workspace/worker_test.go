// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"gitlab.com/blockchain-privacy/dakar/analytics/graph"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
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

	opt := workspace.TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: new(int64(1))},
		InputRange:  &workspace.AmountRange{Min: new(int64(1000000)), Max: new(int64(10000000))},
		OutputRange: &workspace.AmountRange{Min: new(int64(1)), Max: new(int64(10000000))},
	}

	m := NewMutex()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// insert 3 selectors into db
	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		constants.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		constants.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	_, _, err = AddSelector(ctx, dbHandle, m, opt,
		constants.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	wrapper := graph.NewWrapper(ctx, dbHandle)
	wrapper.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, status.SetLastClassifiedBlockID(ctx, dbHandle, int64(db.ClassifierFileLastBlock)))
	require.NoError(t, wrapper.LoadGraphs(graph.NewDashConfig()))

	w := NewWorker(NewMutex(), dbHandle, wrapper)
	w.RegisterMetrics(prometheus.NewRegistry())
	w.SetLoopInterval(1)

	wg := sync.WaitGroup{}
	wg.Go(func() { w.Start(ctx) })

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

type counterWork struct {
}

func (p *counterWork) Run(context.Context, *Mutex, external.Database, *graph.Wrapper) error {
	return nil
}

func TestAddWork(t *testing.T) {
	w := NewWorker(NewMutex(), nil, nil)
	w.SetWaitForInMemoryGraph(false)
	w.RegisterMetrics(prometheus.NewRegistry())

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Go(func() { w.Start(ctx) })

	// queue work
	const workQueueItems = 200
	channels := make([]chan error, workQueueItems)
	wg2 := sync.WaitGroup{}
	wg2.Go(func() {
		for i := range workQueueItems {
			ch := w.AddWork(ctx, &counterWork{})
			if ch == nil {
				t.Error("couldn't add work")
				return
			}

			channels[i] = ch
		}
	})

	// wait for queuing to be finished
	wg2.Wait()
	ticker := time.Tick(time.Second * 30)
	for _, ch := range channels {
		select {
		case err := <-ch:
			require.NoError(t, err)
			// wait for all work items to be done
			continue

		case <-ticker:
			// prevent infinite loop in case something went wrong
			t.Error("worker took to long to finish work")
			break
		}
	}

	// stop worker
	cancel()
	wg.Wait()
}
