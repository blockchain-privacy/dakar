// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package classifier

import (
	"backend/blockiterator"
	"backend/db"
	"backend/db/analytics"
	"backend/db/status"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNewClassifier(t *testing.T) {
	classifier := NewClassifier(t.Context(), nil, NewDashConfig())

	require.NotNil(t, classifier)
}

func TestClassifier_IncrementState(t *testing.T) {
	classifier := NewClassifier(t.Context(), nil, NewDashConfig())

	for range 10 {
		classifier.blocksProcessed = 1
		require.NoError(t, classifier.IncrementState())
	}

	require.EqualValues(t, 10, classifier.state.ID)
}

func TestClassifier_Empty(t *testing.T) {
	classifier := NewClassifier(t.Context(), nil, NewDashConfig())

	require.False(t, classifier.Empty())
	// simulate a block having been processed
	classifier.blocksProcessed = 1
	require.NoError(t, classifier.IncrementState())
	require.True(t, classifier.Empty())
}

func TestClassifier_CalculateInitialState(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	classifier := NewClassifier(ctx, nil, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())

	classifier.db = dbHandle

	// status not set yet
	require.Error(t, classifier.CalculateInitialState(ctx))

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  new(true),
		LastBlockID: new(int64(5)),
	}))

	require.NoError(t, classifier.CalculateInitialState(ctx))
	require.EqualValues(t, 5, classifier.state.Top)
	require.EqualValues(t, 1, classifier.state.ID)
}

func TestClassifier_NextBlock(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  new(false),
		LastBlockID: new(int64(db.BlockFileLastBlock)),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())

	// set to first available block
	classifier.state.ID = db.BlockFileFirstBlock
	classifier.state.Top = db.BlockFileFirstBlock

	got, err := classifier.Next(ctx)
	require.NoError(t, err)
	require.True(t, got)
	require.EqualValues(t, db.BlockFileLastBlock, classifier.state.Top)
}

func TestClassifier_Props(t *testing.T) {
	classifier := NewClassifier(t.Context(), nil, NewDashConfig())

	require.NotEmpty(t, classifier.Props())
}

func TestClassifier_Iterate(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: new(false),
		// first block of the file
		LastBlockID: new(int64(db.ClassifierFileFirstBlock)),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())
	// state is set to block 0, which does not exist in database
	_, err := classifier.Iterate(ctx)
	require.Error(t, err)

	classifier.state.ID = db.ClassifierFileFirstBlock
	classifier.state.Top = db.ClassifierFileFirstBlock

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	_, err = classifier.Iterate(ctx)
	require.NoError(t, err)

	// check mixing count after classification
	mixingCount, _, _, _, _, err := analytics.GetDashTransactionTypeCount(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, mixingCount)
}

func TestMultipleBlockIteration(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	fileBlockCount := int64(db.ClassifierFileLastBlock - db.ClassifierFileFirstBlock)

	ctx, cancelFunc := db.GetTaskContext()
	defer cancelFunc()

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))
	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  new(false),
		LastBlockID: new(int64(db.ClassifierFileLastBlock)),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying:         new(false),
		LastClassifiedBlockID: new(int64(db.ClassifierFileFirstBlock)),
	}))

	classifier2 := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier2.RegisterMetrics(prometheus.NewRegistry())

	classifier2.state.ID = db.ClassifierFileFirstBlock
	classifier2.state.Top = db.ClassifierFileLastBlock
	var numIteratedBlocks int64
	require.NoError(t, blockiterator.StartIteration(classifier2, time.Second*10, func() {
		numIteratedBlocks += classifier2.Props().ProcessedBlockCount
		if numIteratedBlocks >= fileBlockCount {
			cancelFunc()
		}
	}))
}

func TestClassifier_PostExecution(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())

	require.NoError(t, classifier.PostExecution(ctx))
}

func Test_setInitialClassifierID(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying:         new(true),
		LastClassifiedBlockID: new(int64(700)),
	}))
	require.NoError(t, setInitialClassifierID(ctx, dbHandle, 0))
}

func TestBlockIterator(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: new(false),
		// let's classify 2 blocks
		LastBlockID: new(int64(db.ClassifierFileFirstBlock + 3)),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying: new(false),
		// let's classify 3 blocks
		LastClassifiedBlockID: new(int64(db.ClassifierFileFirstBlock)),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()
	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())

	numIteratedBlocks := int64(0)
	require.NoError(t, blockiterator.StartIteration(classifier, time.Second*10, func() {
		numIteratedBlocks += classifier.Props().ProcessedBlockCount
		if numIteratedBlocks >= 3 {
			cancelFunc()
		}
	}))
}

func TestBlockIteratorImmediateExit(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: new(false),
		// let's classify 2 blocks
		LastBlockID: new(int64(db.ClassifierFileFirstBlock + 3)),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying: new(false),
		// let's classify 3 blocks
		LastClassifiedBlockID: new(int64(db.ClassifierFileFirstBlock)),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	// immediately cancel
	cancelFunc()
	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, blockiterator.StartIteration(classifier, time.Second*10, nil))
}
