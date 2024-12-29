package db

import (
	"backend/external"
	"backend/testhelper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v240"
	"github.com/qrest/gomisc/serror"
	"log"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/stretchr/testify/require"
)

const (
	// maxRetries is the number of transaction retries in case of an error response
	maxRetries = 5
	// retrySleepDuration is the duration between retries
	retrySleepDuration = time.Second * 5
)

var (
	// ErrBlockNotFound is returned if no block was found
	ErrBlockNotFound = errors.New("no block found")
	// ErrTransactionNotFound is returned if a requested transaction has not been found
	ErrTransactionNotFound = errors.New("no transaction found")
	// ErrAddressNotFound is returned if no address has been found
	ErrAddressNotFound        = errors.New("no address found")
	ErrEmptyRequestArgument   = errors.New("received empty argument")
	ErrInvalidRequestArgument = errors.New("received invalid argument")
	errInvalidResult          = errors.New("invalid result")
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// UIDNode holds the uid of a database node. Useful for connecting entities.
type UIDNode struct {
	UID string `json:"uid,omitempty"`
}

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "database"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

// GetLongTaskContext returns a context with a timeout of 2 hours
func GetLongTaskContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Hour*2)
}

// GetTaskContext returns a context with a timeout of 20 minutes
func GetTaskContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Minute*20)
}

// GetShortTaskContext returns a context with a timeout of 10 seconds
func GetShortTaskContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*10)
}

func AddShortTaskContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Second*10)
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
func ExecTx(ctx context.Context, tx *dgo.Txn, req *api.Request) (*api.Response, error) {
	if req == nil || tx == nil {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	resp, err := tx.Do(ctx, req)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
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

// DropAll drops ALL data from the database, schema included
func DropAll(db external.Database) error {
	ctx, cancel := GetTaskContext()
	defer cancel()
	err := db.Alter(ctx, &api.Operation{
		DropAll: true,
	})
	if err != nil {
		return serror.New(err)
	}

	return nil
}

// CreateCommaList returns a formatted string which contains all given uids for usage with Dgraph.
// // Example: 0x123,0x1a1d
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

// CreateCommaListQuotationMarks returns a formatted string which contains all given uids for usage with Dgraph.
// // Each given string is put in quotation marks
// // Example: "0x123","0x1a1d"
func CreateCommaListQuotationMarks(uids []string) string {
	var uidEnum string
	for i, uid := range uids {
		uidEnum += "\"" + uid + "\""
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
	if !database.IsDirty.Load() && database.FileNameKey == fileKey {
		return
	}

	ctx, cancel := GetTaskContext()
	defer cancel()

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
	case testhelper.UseBTCPrivacyFile:
		fileBytes = testhelper.BTCPrivacyFile
	default:
		log.Panic("invalid file key")
	}

	if err := InsertArbitraryJSON(ctx, database, fileBytes); err != nil {
		log.Panic("could not upsert block data", err)
		return
	}

	database.IsDirty.Store(false)
	database.FileNameKey = fileKey
}

// SetupDBWithoutData returns the database to its initial state:
// drops ALL data and sets up the schema
func SetupDBWithoutData(t *testing.T, database *testhelper.TestDB) {
	// reset db
	require.NoError(t, DropAll(database))

	// set up schema
	require.NoError(t, SetupSchema(database))

	database.IsDirty.Store(true)
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
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
	v, ok := resp.GetMetrics().GetNumUids()["mutation_cost"]
	return ok && v > 0
}
