package classifier

import (
	"backend/blockiterator"
	"backend/db"
	"backend/db/analytics"
	"backend/db/status"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestNewClassifier(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, NewDashConfig())

	require.NotNil(t, classifier)
}

func TestClassifier_IncrementState(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, NewDashConfig())

	for range 10 {
		classifier.blocksProcessed = 1
		require.NoError(t, classifier.IncrementState())
	}

	require.EqualValues(t, 10, classifier.state.ID)
}

func TestClassifier_Empty(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, NewDashConfig())

	require.False(t, classifier.Empty())
	// simulate a block having been processed
	classifier.blocksProcessed = 1
	require.NoError(t, classifier.IncrementState())
	require.True(t, classifier.Empty())
}

func TestClassifier_CalculateInitialState(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	classifier := NewClassifier(ctx, nil, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())

	classifier.db = dbHandle

	// status not set yet
	require.Error(t, classifier.CalculateInitialState(ctx))

	yes := true
	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  &yes,
		LastBlockID: testhelper.GetPointer[int64](5),
	}))

	require.NoError(t, classifier.CalculateInitialState(ctx))
	require.EqualValues(t, 5, classifier.state.Top)
	require.EqualValues(t, 1, classifier.state.ID)
}

func TestClassifier_NextBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	no := false
	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  &no,
		LastBlockID: testhelper.GetPointer[int64](testhelper.BlockFileLastBlock),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())

	// set to first available block
	classifier.state.ID = testhelper.BlockFileFirstBlock
	classifier.state.Top = testhelper.BlockFileFirstBlock

	got, err := classifier.Next(ctx)
	require.NoError(t, err)
	require.True(t, got)
	require.EqualValues(t, testhelper.BlockFileLastBlock, classifier.state.Top)
}

func TestClassifier_Props(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, NewDashConfig())

	require.NotEmpty(t, classifier.Props())
}

func TestClassifier_Iterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: testhelper.GetPointer(false),
		// first block of the file
		LastBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())
	// state is set to block 0, which does not exist in database
	_, err := classifier.Iterate(ctx)
	require.Error(t, err)

	classifier.state.ID = testhelper.ClassifierFileFirstBlock
	classifier.state.Top = testhelper.ClassifierFileFirstBlock

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	_, err = classifier.Iterate(ctx)
	require.NoError(t, err)

	// check mixing count after classification
	mixingCount, _, _, _, _, err := analytics.GetDashTransactionTypeCount(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, mixingCount)
}

func TestMultipleBlockIteration(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	fileBlockCount := int64(testhelper.ClassifierFileLastBlock - testhelper.ClassifierFileFirstBlock)

	ctx, cancelFunc := db.GetShortTaskContext()
	defer cancelFunc()

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))
	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling:  testhelper.GetPointer(false),
		LastBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileLastBlock),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying:         testhelper.GetPointer(false),
		LastClassifiedBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock),
	}))

	classifier2 := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier2.RegisterMetrics(prometheus.NewRegistry())

	classifier2.state.ID = testhelper.ClassifierFileFirstBlock
	classifier2.state.Top = testhelper.ClassifierFileLastBlock
	var numIteratedBlocks int64
	require.NoError(t, blockiterator.StartIteration(classifier2, time.Second*10, func() {
		numIteratedBlocks += classifier2.Props().ProcessedBlockCount
		if numIteratedBlocks >= fileBlockCount {
			cancelFunc()
		}
	}))
}

func TestClassifier_PostExecution(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())

	require.NoError(t, classifier.PostExecution(ctx))
}

func Test_setInitialClassifierID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	yes := true
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying:         &yes,
		LastClassifiedBlockID: testhelper.GetPointer[int64](700),
	}))
	require.NoError(t, setInitialClassifierID(ctx, dbHandle))
}

func TestBlockIterator(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: testhelper.GetPointer(false),
		// let's classify 2 blocks
		LastBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock + 3),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying: testhelper.GetPointer(false),
		// let's classify 3 blocks
		LastClassifiedBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock),
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
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, status.SetCrawlerStatus(ctx, dbHandle, status.CrawlerStatus{
		IsCrawling: testhelper.GetPointer(false),
		// let's classify 2 blocks
		LastBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock + 3),
	}))
	require.NoError(t, status.SetClassifierStatus(ctx, dbHandle, status.ClassifierStatus{
		IsClassifying: testhelper.GetPointer(false),
		// let's classify 3 blocks
		LastClassifiedBlockID: testhelper.GetPointer[int64](testhelper.ClassifierFileFirstBlock),
	}))

	ctx, cancelFunc := db.GetShortTaskContext()
	// immediatly cancel
	cancelFunc()
	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	classifier.RegisterMetrics(prometheus.NewRegistry())

	require.NoError(t, blockiterator.StartIteration(classifier, time.Second*10, nil))
}
