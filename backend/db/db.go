package db

import (
	"backend/external"
	"backend/testhelper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qrest/gomisc/serror"
	"log"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/stretchr/testify/require"
)

const (
	// backendTimeout is the duration until a request originating from the backend times out
	backendTimeout = time.Minute * 20
	// maxRetries is the number of transaction retries in case of an error response
	maxRetries = 5
	// retrySleepDuration is the duration between retries
	retrySleepDuration = time.Second * 5
)

var thisLogger *slog.Logger

var (
	// ErrBlockNotFound is returned if no block was found
	ErrBlockNotFound = errors.New("no block found")
	// ErrTransactionNotFound is returned if a requested transaction has not been found
	ErrTransactionNotFound = errors.New("no transaction found")
	// ErrAddressNotFound is returned if no address has been found
	ErrAddressNotFound        = errors.New("no address found")
	ErrEmptyRequestArgument   = errors.New("received empty argument")
	ErrInvalidRequestArgument = errors.New("received invalid argument")
	errInvalidTimeout         = errors.New("invalid timeout")
	errInvalidResult          = errors.New("invalid result")
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// UIDNode holds the uid of a database node. Useful for connecting entities.
type UIDNode struct {
	UID string `json:"uid,omitempty"`
}

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	thisLogger = slog.With(slog.String("module", "database"))
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	serror.Log(thisLogger, err, v...)
}

// GetBackendContext returns a context with a runtime of backendTimeout and a cancel function
func GetBackendContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), backendTimeout)
}

// execRequest executes the given request
func execRequest(db external.Database, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
	if timeoutPerRequest <= 0 {
		return nil, serror.New(errInvalidTimeout)
	}

	if req == nil {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()

	resp, err := db.Mutate(ctx, req)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
}

// execReadOnlyRequest executes the given request, vars is allowed to be nil
func execReadOnlyRequest(db external.Database, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	if timeoutPerRequest <= 0 {
		return nil, serror.New(errInvalidTimeout)
	}

	if q == "" {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()

	resp, err := db.Query(ctx, q, vars)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
}

// WithRetry calls the given function. If dgo.ErrAborted is returned, the function
// is called a few more times. Between each call retryDuration is waited.
func WithRetry(f func() error, retryDuration time.Duration) error {
	var err error
	var encounteredError bool
	for range maxRetries {
		if encounteredError {
			// Retry the transaction if it was aborted
			warn(fmt.Errorf("encountered error, retrying: %w", err))
			time.Sleep(retryDuration)
		}

		if err = f(); errors.Is(err, dgo.ErrAborted) {
			encounteredError = true
			continue
		}

		break
	}

	if encounteredError && err == nil {
		info("retryed transaction was successful")
	}

	return err
}

// ExecTx executes the given request. The caller is responsible for
// retrying the transactions in case it is discarded (check error for dgo.ErrAborted).
func ExecTx(tx *dgo.Txn, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
	if timeoutPerRequest <= 0 {
		return nil, serror.New(errInvalidTimeout)
	}

	if req == nil || tx == nil {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutPerRequest)
	defer cancel()

	resp, err := tx.Do(ctx, req)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
}

// TxWithRetry executes the given request. In case the request fails repeat it
func TxWithRetry(db external.Database, timeoutPerRequest time.Duration, req *api.Request) error {
	_, err := TxWithRetryAndResponse(db, timeoutPerRequest, req)
	return err
}

// TxWithRetryAndResponse executes the given request. In case the request fails repeat it
func TxWithRetryAndResponse(db external.Database, timeoutPerRequest time.Duration,
	req *api.Request) (*api.Response, error) {
	var resp *api.Response
	var err error

	err = WithRetry(func() error {
		resp, err = execRequest(db, timeoutPerRequest, req)
		return err
	}, retrySleepDuration)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// MutationWithRetry executes the given request. In case the request fails repeat it
func MutationWithRetry(ctx context.Context, db external.Database, req *api.Request) error {
	_, err := MutationWithRetryAndResponse(ctx, db, req)
	return err
}

// MutationWithRetryAndResponse executes the given request. In case the request fails repeat it
func MutationWithRetryAndResponse(ctx context.Context, db external.Database,
	req *api.Request) (*api.Response, error) {
	var resp *api.Response
	var err error

	err = WithRetry(func() error {
		resp, err = db.Mutate(ctx, req)
		return err
	}, retrySleepDuration)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// ReadOnlyTxVarWithRetry executes the given request. In case the request fails repeats it
func ReadOnlyTxVarWithRetry(db external.Database, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	var resp *api.Response
	var err error

	err = WithRetry(func() error {
		resp, err = execReadOnlyRequest(db, timeoutPerRequest, q, vars)
		return err
	}, retrySleepDuration)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// QueryVarWithRetry executes the given request. In case the request fails repeats it
func QueryVarWithRetry(ctx context.Context, db external.Database, q string,
	vars map[string]string) (*api.Response, error) {
	var resp *api.Response
	var err error

	err = WithRetry(func() error {
		resp, err = db.Query(ctx, q, vars)
		return err
	}, retrySleepDuration)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// ReadOnlyTxWithRetry executes the given request. In case the request fails repeats it
func ReadOnlyTxWithRetry(db external.Database, timeoutPerRequest time.Duration, q string) (*api.Response, error) {
	return ReadOnlyTxVarWithRetry(db, timeoutPerRequest, q, nil)
}

// DropAll drops ALL data from the database, schema included
func DropAll(db external.Database) error {
	ctx, cancel := GetBackendContext()
	defer cancel()
	err := db.Alter(ctx, &api.Operation{
		DropAll: true,
	})
	if err != nil {
		return serror.New(err)
	}

	return nil
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

// returns true if the given input does not contain special characters
func isValidQueryInput(input string) bool {
	return !strings.ContainsAny(input, ";,():{}\"'.^`")
}

// CreateCommaArray returns a formatted string which contains all given uids for usage with Dgraph
// Example: [0x123,0x1a1d]
func CreateCommaArray(uids []string) string {
	return "[" + CreateCommaList(uids) + "]"
}

// SetupDB returns the database to its initial state: drops ALL data,
// sets up the schema and inserts data from the provided file
func SetupDB(t *testing.T, database *testhelper.TestDB, fileKey string) {
	// check if database state has been modified. In case it has not, just return
	if !database.IsDirty && database.FileNameKey == fileKey {
		return
	}

	// reset db
	require.NoError(t, DropAll(database))

	// set up schema
	require.NoError(t, SetupSchema(database))

	var fileBytes []byte

	switch fileKey {
	case testhelper.UseClassifierFile:
		fileBytes = testhelper.ClassifierFile
	case testhelper.UseBlockFile:
		fileBytes = testhelper.BlockFile
	case testhelper.UsePrivacyFile:
		fileBytes = testhelper.PrivacyFile
	default:
		log.Panic("invalid file key")
	}

	if err := InsertArbitraryJSON(database, fileBytes); err != nil {
		log.Panic("could not upsert block data", err)
		return
	}

	database.IsDirty = false
	database.FileNameKey = fileKey
}

// SetupDBWithoutData returns the database to its initial state:
// drops ALL data and sets up the schema
func SetupDBWithoutData(t *testing.T, database *testhelper.TestDB) {
	// reset db
	require.NoError(t, DropAll(database))

	// set up schema
	require.NoError(t, SetupSchema(database))

	database.IsDirty = true
}

func GetTypeByUID(ctx context.Context, c external.Database, uid string) (string, error) {
	if uid == "" {
		return "", serror.New(ErrEmptyRequestArgument)
	}

	const query = `query Q($uid:string){
				q(func: uid($uid)){
					dgraph.type
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$uid": uid})
	if err != nil {
		return "", serror.New(err)
	}

	// json struct
	var r struct {
		Type []struct {
			Type []string `json:"dgraph.type,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return "", serror.New(err)
	}

	if len(r.Type) != 1 || len(r.Type[0].Type) != 1 {
		return "", serror.New(errInvalidResult)
	}

	return r.Type[0].Type[0], nil
}

// HasMutationCost returns true if the response has a mutation cost attached.
// This happens if a request mutated data in the database.
func HasMutationCost(resp *api.Response) bool {
	v, ok := resp.Metrics.NumUids["mutation_cost"]
	return ok && v > 0
}
