package db

import (
	"backend/cmd/cliutil"
	"io"

	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
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
	thisLogger.Println(v)
}

func GetBackendContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), backendTimeout)
}

func GetFrontendContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), frontEndTimout)
}

// execTx executes the given request
func execTx(dgraph *dgo.Dgraph, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()
	return dgraph.NewTxn().Do(ctx, req)
}

// TxWithRetry executes the given request. In case the request fails repeat it
func TxWithRetry(dgraph *dgo.Dgraph, timeoutPerRequest time.Duration, req *api.Request) (err error) {
	for i := 0; i < maxRetries; i++ {
		if _, err = execTx(dgraph, timeoutPerRequest, req); err == nil {
			return
		}
		info(cliutil.ShowCallInfo(), "encountered error retrying:", err, "request:", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// Execute the given request. In case the request fails repeat it. Also returns the response
func TxWithRetryAndResponse(dgraph *dgo.Dgraph, ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	for i := 0; i < maxRetries; i++ {
		if resp, err = dgraph.NewTxn().Do(ctx, req); err == nil {
			return
		}
		info(cliutil.ShowCallInfo(), "encountered error retrying:", err, "request:", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// execReadOnlyTxWithVars executes the given request, vars is allowed to be nil
func execReadOnlyTxWithVars(dgraph *dgo.Dgraph, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()

	return dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
}

// ReadOnlyTxVarWithRetryAndTimeout executes the given request. In case the request fails repeats it
func ReadOnlyTxVarWithRetry(dgraph *dgo.Dgraph, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	var err error
	for i := 0; i < maxRetries; i++ {
		resp, txErr := execReadOnlyTxWithVars(dgraph, timeoutPerRequest, q, vars)
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
func ReadOnlyTxWithRetry(dgraph *dgo.Dgraph, timeoutPerRequest time.Duration, q string) (*api.Response, error) {
	return ReadOnlyTxVarWithRetry(dgraph, timeoutPerRequest, q, nil)
}

// drops ALL data from the database, schema included
func DropAll(c *dgo.Dgraph) error {
	ctx, cancel := GetBackendContext()
	defer cancel()
	return c.Alter(ctx, &api.Operation{
		DropOp: api.Operation_ALL,
	})
}

// create a new dgraph client connecting to the specified host and port
func CreateClient(endpoint string) (*dgo.Dgraph, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)))

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return nil, conn, err
	}

	return dgo.NewDgraphClient(api.NewDgraphClient(conn)), conn, nil
}

// gets the number of instances of the given type in the database
func GetCount(c *dgo.Dgraph, dbType string) (count uint64, err error) {
	query := fmt.Sprintf(`{
				 q(func: type(%s)){
					count(uid)
				  }
				}
				`, dbType)

	ctx, cancel := GetBackendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().Query(ctx, query)
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
