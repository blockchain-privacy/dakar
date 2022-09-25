package db

import (
	"backend/cmd/cliutil"
	"backend/external"
	"backend/testhelper"

	"context"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// backendTimeout is the duration until a request originating from the backend times out
	backendTimeout = time.Minute * 20
	// frontEndTimout is the duration until a request originating from the frontend times out
	frontEndTimout = time.Second * 30
	// maxRetries is the number of transaction retries in case of an error response
	maxRetries = 5
	// retrySleepDuration is the duration between retries
	retrySleepDuration = time.Second * 5

	// loggerPrefix is the prefix which is printed for each log message
	loggerPrefix = "\033[0;33mdb\u001B[0m\t"
)

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

// GetBackendContext returns a context with a runtime of backendTimeout and a cancel function
func GetBackendContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), backendTimeout)
}

// GetFrontendContext returns a context with a runtime of frontEndTimout and a cancel function
func GetFrontendContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), frontEndTimout)
}

// execTx executes the given request
func execTx(db external.Database, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()
	return db.Mutate(ctx, req)
}

// execExistingTx executes the given request
func execExistingTx(tx *dgo.Txn, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()
	return tx.Do(ctx, req)
}

// TxWithRetry executes the given request. In case the request fails repeat it
func TxWithRetry(db external.Database, timeoutPerRequest time.Duration, req *api.Request) error {
	_, err := TxWithRetryAndResponse(db, timeoutPerRequest, req)
	return err
}

// TxWithRetryAndResponse executes the given request. In case the request fails repeat it
func TxWithRetryAndResponse(db external.Database, timeoutPerRequest time.Duration,
	req *api.Request) (resp *api.Response, err error) {
	for i := 0; i < maxRetries; i++ {
		if resp, err = execTx(db, timeoutPerRequest, req); err == nil {
			return
		}
		info(cliutil.ShowCallInfo(), "encountered error retrying:", err, "request:", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// ExistingTxWithRetry executes the given request. In case the request fails repeat it
func ExistingTxWithRetry(tx *dgo.Txn, timeoutPerRequest time.Duration, req *api.Request) error {
	_, err := ExistingTxWithRetryAndResponse(tx, timeoutPerRequest, req)
	return err
}

// ExistingTxWithRetryAndResponse executes the given request. In case the request fails repeat it
func ExistingTxWithRetryAndResponse(tx *dgo.Txn, timeoutPerRequest time.Duration,
	req *api.Request) (resp *api.Response, err error) {
	for i := 0; i < maxRetries; i++ {
		if resp, err = execExistingTx(tx, timeoutPerRequest, req); err == nil {
			return
		}
		info(cliutil.ShowCallInfo(), "encountered error retrying:", err, "request:", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// execReadOnlyTx executes the given request, vars is allowed to be nil
func execReadOnlyTx(db external.Database, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()

	return db.Query(ctx, q, vars)
}

// ReadOnlyTxVarWithRetry executes the given request. In case the request fails repeats it
func ReadOnlyTxVarWithRetry(db external.Database, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, txErr := execReadOnlyTx(db, timeoutPerRequest, q, vars)
		if txErr == nil {
			return resp, nil
		}
		err = txErr
		info(cliutil.ShowCallInfo(), "encountered error retrying:", err, "query:", q, "vars:", vars)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return nil, err
}

// ReadOnlyTxWithRetry executes the given request. In case the request fails repeats it
func ReadOnlyTxWithRetry(db external.Database, timeoutPerRequest time.Duration, q string) (*api.Response, error) {
	return ReadOnlyTxVarWithRetry(db, timeoutPerRequest, q, nil)
}

// DropAll drops ALL data from the database, schema included
func DropAll(db external.Database) error {
	ctx, cancel := GetBackendContext()
	defer cancel()
	return db.Alter(ctx, &api.Operation{
		DropAll: true,
	})
}

// CreateClient create a new dgraph client connecting to the specified host and port
func CreateClient(endpoint string) (external.Database, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)))

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return nil, conn, err
	}

	return &external.GraphDB{Dgraph: dgo.NewDgraphClient(api.NewDgraphClient(conn))}, conn, nil
}

// CreateCommaList returns a formatted string which contains all given uids for usage with Dgraph
// Example: 0x123,0x1a1d
func CreateCommaList(uids []string) string {
	var uidEnum string
	for i, uid := range uids {
		uidEnum += uid
		if i+1 < len(uids) {
			uidEnum += ","
		}
	}
	return uidEnum
}

// CreateCommaArray returns a formatted string which contains all given uids for usage with Dgraph
// Example: [0x123,0x1a1d]
func CreateCommaArray(uids []string) string {
	return "[" + CreateCommaList(uids) + "]"
}

// IsConnectionEstablished test the database connection
func IsConnectionEstablished(c external.Database) bool {
	ctx, cancel := GetBackendContext()
	defer cancel()
	response, err := c.Query(ctx, "{q(func: has(Meta.schemaVersion),first:1){uid}}", nil)
	_ = response
	return err == nil
}

// WaitForDatabase waits until the database is ready to receive requests
func WaitForDatabase(db external.Database) bool {
	const maxRetries = 20
	const retrySleepDuration = time.Second * 5

	var printedErrMessage bool

	for i := 0; i < maxRetries; i++ {
		if IsConnectionEstablished(db) {
			if printedErrMessage {
				info("Successfully established connection to database.")
			}
			return true
		}

		if !printedErrMessage {
			info("Waiting for database")
			printedErrMessage = true
		}

		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	info("Database is not ready to receive requests.")

	return false
}

type ContainerName string

const (
	ContainerNameDB        = "dgraph_db"
	ContainerNameStatus    = "dgraph_status"
	ContainerNameUser      = "dgraph_user"
	ContainerNameProcessor = "dgraph_processor"
)

// RunDgraphTests connects to the given dgraph container and runs all tests
// packageDBHandle should be set to the global db interface handle of the package module.
func RunDgraphTests(m *testing.M, packageDBHandle *external.Database, containerName ContainerName) {
	if testhelper.IsCIActive() {
		// create dgraph client
		graphDB, c, err := CreateClient(string(containerName) + ":9080")
		if err != nil {
			log.Panic(err)
			return
		}
		defer func(c *grpc.ClientConn) {
			err := c.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(c)

		if !WaitForDatabase(graphDB) {
			log.Panic("Could not connect to database", err)
			return
		}

		*packageDBHandle = graphDB
	}

	m.Run()
}

// SetupDB returns the database to its initial state: drops ALL data,
// sets up the schema and inserts data from the provided file
func SetupDB(t *testing.T, database external.Database, blockFileName string) {
	// reset db
	require.NoError(t, DropAll(database))

	// set up schema
	require.NoError(t, SetupSchema(database))

	fileBytes, err := os.ReadFile(blockFileName)
	require.NoError(t, err)

	if err := InsertArbitrary(database, fileBytes); err != nil {
		log.Panic("Could not upsert block data", err)
		return
	}

	//// add blocks
	//for _, b := range blocks {
	//	if err := UpsertBlock(database, b); err != nil {
	//		log.Panic("Could not upsert block data", err)
	//		return
	//	}
	//}
}

// SetupDBWithoutData returns the database to its initial state:
// drops ALL data and sets up the schema
func SetupDBWithoutData(t *testing.T, database external.Database) {
	// reset db
	require.NoError(t, DropAll(database))

	// set up schema
	require.NoError(t, SetupSchema(database))
}
