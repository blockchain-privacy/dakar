package db

import (
	"backend/cmd/cliutil"
	"backend/external"

	"context"
	"encoding/json"
	"errors"
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
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)))

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return nil, conn, err
	}

	return &external.GraphDB{Dgraph: dgo.NewDgraphClient(api.NewDgraphClient(conn))}, conn, nil
}

// GetCount gets the number of instances of the given type in the database
func GetCount(db external.Database, dbType string) (count uint64, err error) {
	query := fmt.Sprintf(`{
				 q(func: type(%s)){
					count(uid)
				  }
				}
				`, dbType)

	ctx, cancel := GetBackendContext()
	defer cancel()
	resp, err := db.Query(ctx, query, nil)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		GetCount []struct {
			Count uint64 `json:"count"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.GetCount) != 1 {
		err = errors.New("wrong number of objects returned")
		return
	}

	count = r.GetCount[0].Count

	return
}

// CreateUIDList returns a formatted string which contains all given uids for usage with Dgraph
// Example: [0x123,0x1a1d]
func CreateUIDList(uids []string) string {
	uidList := "["
	for i, uid := range uids {
		uidList += uid
		if i+1 < len(uids) {
			uidList += ","
		}
	}
	uidList += "]"
	return uidList
}
