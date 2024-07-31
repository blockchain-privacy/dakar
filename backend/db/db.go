package db

import (
	"backend/external"
	"backend/testhelper"
	"context"
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
	ErrAddressNotFound      = errors.New("no address found")
	ErrEmptyRequestArgument = errors.New("received empty argument")
	errInvalidTimeout       = errors.New("invalid timeout")
	errInvalidResult        = errors.New("invalid result")
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

// execTx executes the given request
func execTx(db external.Database, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
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

// execReadOnlyTx executes the given request, vars is allowed to be nil
func execReadOnlyTx(db external.Database, timeoutPerRequest time.Duration, q string,
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

// execExistingTx executes the given request
func execExistingTx(tx *dgo.Txn, timeoutPerRequest time.Duration, req *api.Request) (*api.Response, error) {
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
	req *api.Request) (resp *api.Response, err error) {
	for i := range maxRetries {
		if resp, err = execTx(db, timeoutPerRequest, req); err == nil || !errors.Is(err, dgo.ErrAborted) {
			return
		}

		// Retry the transaction if it was aborted
		warn(fmt.Errorf("encountered error, retrying: %w", err), "request", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// MutationWithRetry executes the given request. In case the request fails repeat it
func MutationWithRetry(ctx context.Context, db external.Database, req *api.Request) error {
	_, err := MutationWithRetryAndResponse(ctx, db, req)
	return err
}

// MutationWithRetryAndResponse executes the given request. In case the request fails repeat it
func MutationWithRetryAndResponse(ctx context.Context, db external.Database,
	req *api.Request) (resp *api.Response, err error) {
	for i := range maxRetries {
		if resp, err = db.Mutate(ctx, req); err == nil || !errors.Is(err, dgo.ErrAborted) {
			return
		}

		err = serror.New(err)

		// Retry the transaction if it was aborted
		warn(fmt.Errorf("encountered error, retrying: %w", err), "request", req)
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
	for i := range maxRetries {
		if resp, err = execExistingTx(tx, timeoutPerRequest, req); err == nil || !errors.Is(err, dgo.ErrAborted) {
			return
		}

		// Retry the transaction if it was aborted
		warn(fmt.Errorf("encountered error, retrying: %w", err), "request", req)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
}

// ReadOnlyTxVarWithRetry executes the given request. In case the request fails repeats it
func ReadOnlyTxVarWithRetry(db external.Database, timeoutPerRequest time.Duration, q string,
	vars map[string]string) (*api.Response, error) {
	var err error
	for i := range maxRetries {
		resp, txErr := execReadOnlyTx(db, timeoutPerRequest, q, vars)
		if txErr == nil {
			return resp, nil
		}
		err = txErr

		if errors.Is(err, errInvalidTimeout) || errors.Is(err, ErrEmptyRequestArgument) {
			return nil, err
		}

		warn(fmt.Errorf("encountered error, retrying: %w", err), "query", q, "vars", vars)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return nil, err
}

// QueryVarWithRetry executes the given request. In case the request fails repeats it
func QueryVarWithRetry(ctx context.Context, db external.Database, q string,
	vars map[string]string) (resp *api.Response, err error) {
	for i := range maxRetries {
		resp, err = db.Query(ctx, q, vars)
		if err == nil {
			return
		}
		err = serror.New(err)

		warn(fmt.Errorf("encountered error, retrying: %w", err), "query", q, "vars", vars)
		if i+1 < maxRetries {
			time.Sleep(retrySleepDuration)
		}
	}

	return
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
