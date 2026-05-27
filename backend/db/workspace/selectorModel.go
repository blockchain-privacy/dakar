// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"slices"
	"time"

	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
)

const SelectorDType = "Selector"

const (
	// StatusWaiting is set when a selector has been created and is not yet run
	StatusWaiting = "waiting"
	// StatusError is set when an error occurred while a selector has been run
	StatusError = "error"
	// StatusSuccess is set when no error occurred while a selector has been run
	StatusSuccess = "success"
	// MaxResultsPerSelector is the maximum allowed results per selector
	MaxResultsPerSelector = 20_000
	// ErrorCodeResultLimitExceeded is set when the selector returned more than MaxConnectionsPerSelector
	ErrorCodeResultLimitExceeded = "result_limit_exceeded"
)

const selectorMaxItems = 200

var validTypes = map[string]bool{constants.TypeTxProp: true, constants.TypeTxGraph: true, constants.TypeHeuristic: true}

var validStates = map[string]bool{StatusWaiting: true, StatusError: true, StatusSuccess: true}

// IsTypeValid returns true if the provided type is valid
func IsTypeValid(t string) bool {
	return validTypes[t]
}

// IsStatusValid returns true if the provided status is valid
func IsStatusValid(s string) bool {
	return validStates[s]
}

type Selector struct {
	UID              string      `json:"uid,omitempty"`
	Created          string      `json:"Selector.created,omitempty"`
	Modified         string      `json:"Selector.modified,omitempty"`
	Type             string      `json:"Selector.type,omitempty"`
	Status           string      `json:"Selector.status,omitempty"`
	ErrorCode        string      `json:"Selector.errorCode,omitempty"`
	Parent           *db.UIDNode `json:"Selector.parent,omitempty"`
	Options          string      `json:"Selector.options,omitempty"`
	Results          []any       `json:"Selector.results,omitempty"`
	TotalResultCount *int        `json:"Selector.totalResultCount,omitempty"`
	DType            []string    `json:"dgraph.type,omitempty"`
}

func (s *Selector) SetDType() {
	s.DType = []string{SelectorDType}
}

func (s *Selector) IsValid() bool {
	return s.Options != "" && IsTypeValid(s.Type) && IsStatusValid(s.Status)
}

type AmountRange struct {
	Min *int64 `json:"min,omitempty"`
	Max *int64 `json:"max,omitempty"`
}

func (a AmountRange) IsValid() bool {
	return (a.Min != nil || a.Max != nil) && (a.Max == nil || a.Min == nil || *a.Min <= *a.Max)
}

type Options interface {
	// IsValid returns true if the Options are valid.
	IsValid(ctx context.Context, c external.Database, parentUID string) bool
}

type TxPropOptions struct {
	// StartDate is the start of the time range selection
	StartDate *time.Time `json:"startDate,omitempty"`
	// EndDate is the end of the time range selection
	EndDate *time.Time `json:"endDate,omitempty"`
	// MaxItems is the maximum number of items the selector stores. Can not be higher than selectorMaxItems
	MaxItems *int `json:"maxItems,omitempty"`
	// TransactionTypes contains the transaction types which are included in the selection
	TransactionTypes []string `json:"txTypes,omitempty"`
	// ExcludePrivacyTransactions determines if all transactions with a transaction type should be excluded
	ExcludePrivacyTransactions *bool `json:"excludePrivacyTransactions,omitempty"`
	// InputSum is an amount range of the summed up inputs each transaction has to fulfill
	InputSum *AmountRange `json:"inputSum,omitempty"`
	// OutputSum is an amount range of the summed up outputs each transaction has to fulfill
	OutputSum *AmountRange `json:"outputSum,omitempty"`
	// InputRange is an amount range at least one input of each transaction has to fulfill
	InputRange *AmountRange `json:"inputRange,omitempty"`
	// OutputRange  is an amount range at least one output of each transaction has to fulfill
	OutputRange *AmountRange `json:"outputRange,omitempty"`
}

//nolint:gocyclo
func (o TxPropOptions) IsValid(ctx context.Context, c external.Database, parentUID string) bool {
	// if maxItems is set, it has to be in a valid range
	if o.MaxItems != nil && (*o.MaxItems <= 0 || *o.MaxItems > selectorMaxItems) {
		return false
	}

	if parentUID == "" {
		// both dates must be set
		if o.StartDate == nil || o.EndDate == nil {
			return false
		}

		// start date must be before end date
		if o.StartDate.After(*o.EndDate) {
			return false
		}
		// duration between must be smaller than 60 days (for performance)
		if o.EndDate.Sub(*o.StartDate) > time.Hour*24*60 {
			return false
		}
	}

	// at least one option must be set
	if o.OutputSum == nil && o.InputSum == nil &&
		o.InputRange == nil && o.OutputRange == nil &&
		o.TransactionTypes == nil && o.ExcludePrivacyTransactions == nil {
		return false
	}

	if o.InputSum != nil && !o.InputSum.IsValid() {
		return false
	}

	if o.OutputSum != nil && !o.OutputSum.IsValid() {
		return false
	}

	if o.InputRange != nil && !o.InputRange.IsValid() {
		return false
	}

	if o.OutputRange != nil && !o.OutputRange.IsValid() {
		return false
	}

	// can not exclude all classified transactions and at the same time filter for classified transactions
	if o.TransactionTypes != nil && o.ExcludePrivacyTransactions != nil && *o.ExcludePrivacyTransactions {
		return false
	}

	for _, transactionType := range o.TransactionTypes {
		if !constants.IsTransactionType(transactionType) {
			return false
		}
	}

	if parentUID != "" {
		databaseType, err := db.GetTypeByUID(ctx, c, parentUID)
		if err != nil || databaseType == "" {
			return false
		}

		// if a parent is present, it must be a selector
		if !isParentTypeValid([]string{SelectorDType}, []string{databaseType}) {
			return false
		}
	}

	return true
}

// returns true if any item of parentTypes is in allowedParents
func isParentTypeValid(allowedParents []string, parentTypes []string) bool {
	if len(allowedParents) == 0 {
		return false
	}

	for _, parentType := range parentTypes {
		if slices.Contains(allowedParents, parentType) {
			return true
		}
	}
	return false
}

type TxGraphOptions struct {
	// MaxItems is the maximum number of items the selector stores. Can not be higher than selectorMaxItems
	MaxItems *int `json:"maxItems,omitempty"`
	// IsForward determines the direction of the graph lookup: forward or backward
	IsForward bool `json:"isForward,omitempty"`
	// Depth determines how many levels the transaction graph will be traversed
	Depth *int `json:"depth,omitempty"`
	// ExcludePrivacyTransactions determines if classified transactions should be traversed
	ExcludePrivacyTransactions bool `json:"excludePrivacyTransactions,omitempty"`
}

func (o TxGraphOptions) IsValid(ctx context.Context, c external.Database, parentUID string) bool {
	if parentUID == "" {
		return false
	}

	// if maxItems is set, it has to be in a valid range
	if o.MaxItems != nil && (*o.MaxItems <= 0 || *o.MaxItems > selectorMaxItems) {
		return false
	}

	if o.MaxItems == nil || o.Depth == nil || *o.Depth <= 0 || *o.Depth > 5 {
		return false
	}

	databaseType, err := db.GetTypeByUID(ctx, c, parentUID)
	if err != nil || databaseType == "" {
		return false
	}

	// Dgraph type must be 'Transaction'
	return isParentTypeValid([]string{db.TransactionDType}, []string{databaseType})
}

type WorkItem struct {
	UserUID         string
	WorkspaceUID    string
	SelectorUID     string
	SelectorType    string
	SelectorOptions string
	ParentUID       string
}

type TransactionWithTimestamp struct {
	Hash      string `json:"txhash,omitempty"`
	Timestamp string `json:"ts,omitempty"`
	Type      string `json:"txtype,omitempty"`
	Cluster   *int   `json:"cluster,omitempty"`
}

type Attribution struct {
	Tag      string `json:"tag,omitempty"`
	IsPublic bool   `json:"isPublic"`
}
