package workspace

import (
	"backend/db"
	"time"
)

const SelectorDType = "Selector"

const (
	StatusWaiting = "waiting"
	StatusError   = "error"
	StatusSuccess = "success"

	TypeTransactionProperties = "transactionProperties"
	TypeHeuristic             = "heuristic"

	PrivacyTypeOrigin             = "origin"
	PrivacyTypeMixing             = "mixing"
	PrivacyTypeDestination        = "destination"
	PrivacyTypeCollateralCreation = "cc"
	PrivacyTypeCollateralPayment  = "cp"
)

var validTypes = map[string]bool{TypeTransactionProperties: true, TypeHeuristic: true}

var validStates = map[string]bool{StatusWaiting: true, StatusError: true, StatusSuccess: true}

var validPrivacyTypes = map[string]bool{PrivacyTypeOrigin: true, PrivacyTypeMixing: true,
	PrivacyTypeDestination: true, PrivacyTypeCollateralCreation: true, PrivacyTypeCollateralPayment: true}

// IsTypeValid returns true if the provided type is valid
func IsTypeValid(t string) bool {
	return validTypes[t]
}

// IsStatusValid returns true if the provided status is valid
func IsStatusValid(s string) bool {
	return validStates[s]
}

// IsPrivacyTypeValid returns true if the provided privacy type is valid
func IsPrivacyTypeValid(s string) bool {
	return validPrivacyTypes[s]
}

type Selector struct {
	UID      string      `json:"uid,omitempty"`
	Created  string      `json:"Selector.created,omitempty"`
	Modified string      `json:"Selector.modified,omitempty"`
	Type     string      `json:"Selector.type,omitempty"`
	Status   string      `json:"Selector.status,omitempty"`
	Parent   *db.UIDNode `json:"Selector.parent,omitempty"`
	Options  string      `json:"Selector.options,omitempty"`
	Results  []any       `json:"Selector.results,omitempty"`
	DType    []string    `json:"dgraph.type,omitempty"`
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
	return !((a.Min == nil && a.Max == nil) || (a.Max != nil && a.Min != nil && *a.Min > *a.Max))
}

type Options struct {
	// StartDate is the start of the time range selection
	StartDate *time.Time `json:"startDate,omitempty"`
	// EndDate is the end of the time range selection
	EndDate *time.Time `json:"endDate,omitempty"`
	// PrivacyTypes contains the privacy types which are included in the selection
	PrivacyTypes []string `json:"privacyTypes,omitempty"`
	// ExcludePrivacyTransactions determines if all transactions with a privacy type should be excluded
	ExcludePrivacyTransactions *bool `json:"excludePrivacyTransactions,omitempty"`
	// InputSum is an amount range of the summed up inputs each transaction has to fullfil
	InputSum *AmountRange `json:"inputSum,omitempty"`
	// OutputSum is an amount range of the summed up outputs each transaction has to fullfil
	OutputSum *AmountRange `json:"outputSum,omitempty"`
	// InputRange is an amount range at least one input of each transaction has to fullfil
	InputRange *AmountRange `json:"inputRange,omitempty"`
	// OutputRange  is an amount range at least one output of each transaction has to fullfil
	OutputRange *AmountRange `json:"outputRange,omitempty"`
}

func (o Options) IsValid(hasParent bool) bool {
	if !hasParent {
		// both dates must be set
		if o.StartDate == nil || o.EndDate == nil {
			return false
		}

		// start date must be smaller than end date
		if o.StartDate.Compare(*o.EndDate) > 0 {
			return false
		}
	}

	// at least one option must be set
	if o.OutputSum == nil && o.InputSum == nil &&
		o.InputRange == nil && o.OutputRange == nil &&
		o.PrivacyTypes == nil && o.ExcludePrivacyTransactions == nil {
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

	// can not exclude all privacy transactions and at the same time filter for privacy transactions
	if o.PrivacyTypes != nil && o.ExcludePrivacyTransactions != nil && *o.ExcludePrivacyTransactions {
		return false
	}

	// there are only 5 privacy types
	if o.PrivacyTypes != nil && len(o.PrivacyTypes) > 5 {
		return false
	}

	for _, privacyType := range o.PrivacyTypes {
		if !IsPrivacyTypeValid(privacyType) {
			return false
		}
	}

	return true
}

type WorkItem struct {
	UserUID         string
	WorkspaceUID    string
	SelectorUID     string
	SelectorType    string
	SelectorOptions string
	ParentUID       string
}

type FrontendSelectorResults struct {
	// set if not a heuristic
	Transactions []TransactionWithTimestamp `json:"transactions,omitempty"`
	// set if a heuristic
	Clusters []HeuristicCluster `json:"clusters,omitempty"`
}

type TransactionWithTimestamp struct {
	Timestamp string `json:"ts,omitempty"`
	Hash      string `json:"txhash,omitempty"`
}

type HeuristicCluster struct {
	Transactions []TransactionWithTimestamp `json:"transactions,omitempty"`
	Attributions []Attribution              `json:"attributions,omitempty"`
}

type Attribution struct {
	Tag      string `json:"tag,omitempty"`
	IsPublic bool   `json:"isPublic"`
}
