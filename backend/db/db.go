// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/blockchain-privacy/dakar/external"
	"log/slog"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"gitlab.com/blockchain-privacy/gomisc/serror"

	"github.com/dgraph-io/dgo/v250/protos/api"
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
	ErrTransactionNotFound    = errors.New("no transaction found")
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

// GetShortTaskContext returns a context with a timeout of 1 minute
func GetShortTaskContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Minute)
}

func AddShortTaskContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Minute)
}

func AddTaskContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Minute*20)
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

		if err = f(); err != nil && (errors.Is(err, dgo.ErrAborted) ||
			strings.Contains(err.Error(), "errIndexingInProgress")) {
			encounteredError = true
			continue
		}

		break
	}

	if encounteredError && err == nil {
		info("retried transaction was successful")
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

// CreateCommaList returns a formatted string which contains all given uids for usage with Dgraph.
// // Example: 0x123,0x1a1d
func CreateCommaList(uids []string) string {
	uidEnum := strings.Builder{}
	for i, uid := range uids {
		uidEnum.WriteString(uid)
		if i+1 < len(uids) {
			uidEnum.WriteRune(',')
		}
	}
	return uidEnum.String()
}

// CreateCommaListQuotationMarks returns a formatted string which contains all given uids for usage with Dgraph.
// // Each given string is put in quotation marks
// // Example: "0x123","0x1a1d"
func CreateCommaListQuotationMarks(uids []string) string {
	uidEnum := strings.Builder{}
	for i, uid := range uids {
		uidEnum.WriteString("\"" + uid + "\"")
		if i+1 < len(uids) {
			uidEnum.WriteRune(',')
		}
	}
	return uidEnum.String()
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
