package db

import (
	"backend/cmd/cliutil"
	"backend/external"
	"encoding/json"
	"errors"
	"flag"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
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

// RunDgraphTests creates a dgraph docker container, runs all tests of the calling
// package and removes the container afterwards.
// packageDBHandle should be set to the global db interface handle of the package module.
// The spawned container is removed after the tests are done or after 3 minutes.
func RunDgraphTests(m *testing.M, packageDBHandle *external.Database) {
	// getRandomPortOffset returns a cryptographically UNSAFE integer between 1 and 50.
	getRandomPortOffset := func() int {
		rand.Seed(time.Now().UnixNano())
		return rand.Intn(50-1) + 1 //nolint:gosec
	}

	flag.Parse()
	if testing.Short() {
		return
	}

	var offset = 0
	const anchorPort = 20000

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// try to create container; several tries might be needed because chosen port is already in use
	var resource *dockertest.Resource
	if err := pool.Retry(func() error {
		// generate random port offset
		offset += getRandomPortOffset()
		// pulls an image, creates a container based on it and runs it
		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository:   "dgraph/standalone",
			Tag:          "v21.03.2",
			ExposedPorts: []string{"9080"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"9080": {{HostIP: "0.0.0.0", HostPort: strconv.Itoa(anchorPort + offset)}},
			},
		})
		return err
	}); err != nil {
		log.Panicf("Could not start resource: %s", err)
	}

	// try to set container to expire after 3 minutes
	_ = resource.Expire(180)

	// You can't defer this because os.Exit doesn't care for defer
	defer func(pool *dockertest.Pool, r *dockertest.Resource) {
		err := pool.Purge(r)
		if err != nil {
			log.Fatal(err)
		}
	}(pool, resource)

	hostName := os.Getenv("YOUR_APP_DB_HOST")
	// create dgraph client
	graphDB, c, err := CreateClient(hostName + ":" + strconv.Itoa(anchorPort+offset))
	if err != nil {
		info(err)
		return
	}
	defer func(c *grpc.ClientConn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		if IsConnectionEstablished(graphDB) {
			return nil
		}

		return errors.New("database not ready yet")
	}); err != nil {
		log.Panicf("Could not connect to database: %s", err)
	}

	*packageDBHandle = graphDB
	m.Run()
}

func RunDgraphTestsWithFilledDatabase(m *testing.M, packageDBHandle *external.Database) {
	// getRandomPortOffset returns a cryptographically UNSAFE integer between 1 and 50.
	getRandomPortOffset := func() int {
		rand.Seed(time.Now().UnixNano())
		return rand.Intn(50-1) + 1 //nolint:gosec
	}

	flag.Parse()
	if testing.Short() {
		return
	}

	var offset = 0
	const anchorPort = 20000

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// try to create container; several tries might be needed because chosen port is already in use
	var resource *dockertest.Resource
	if err := pool.Retry(func() error {
		// generate random port offset
		offset += getRandomPortOffset()
		// pulls an image, creates a container based on it and runs it
		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository:   "dgraph/standalone",
			Tag:          "v21.03.2",
			ExposedPorts: []string{"9080"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"9080": {{HostIP: "0.0.0.0", HostPort: strconv.Itoa(anchorPort + offset)}},
			},
		})
		return err
	}); err != nil {
		log.Panicf("Could not start resource: %s", err)
	}

	// try to set container to expire after 3 minutes
	_ = resource.Expire(180)

	// You can't defer this because os.Exit doesn't care for defer
	defer func(pool *dockertest.Pool, r *dockertest.Resource) {
		err := pool.Purge(r)
		if err != nil {
			log.Fatal(err)
		}
	}(pool, resource)

	hostName := os.Getenv("YOUR_APP_DB_HOST")
	// create dgraph client
	graphDB, c, err := CreateClient(hostName + ":" + strconv.Itoa(anchorPort+offset))
	if err != nil {
		info(err)
		return
	}
	defer func(c *grpc.ClientConn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		if IsConnectionEstablished(graphDB) {
			return nil
		}

		return errors.New("database not ready yet")
	}); err != nil {
		log.Panicf("Could not connect to database: %s", err)
	}

	if err := SetupSchema(graphDB); err != nil {
		log.Panicf("Could not set schema: %s", err)
		return
	}

	const fileName = "testdata/blocks_60000_60020.json"
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Panicf("Could not read file: %s, err: %s", fileName, err)
	}
	var blocks []Block

	if err := json.Unmarshal(fileBytes, &blocks); err != nil {
		log.Panicf("Could not unmarshal block data: %s", err)
	}

	for _, b := range blocks {
		if err := UpsertBlock(graphDB, b); err != nil {
			log.Panicf("Could not upsert block data: %s", err)
		}
	}

	*packageDBHandle = graphDB
	m.Run()
}
