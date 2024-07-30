package selectors

import "time"

const DType = "Selector"

var validTypes = map[string]bool{"transactionProperties": true}
var validStates = map[string]bool{"waiting": true, "error": true, "success": true}

// DummyNode holds the uid of a database node
type DummyNode struct {
	UID string `json:"uid,omitempty"`
}

type Selector struct {
	UID      string      `json:"uid,omitempty"`
	Created  string      `json:"Selector.created,omitempty"`
	Modified string      `json:"Selector.modified,omitempty"`
	Type     string      `json:"Selector.type,omitempty"`
	Status   string      `json:"Selector.status,omitempty"`
	Parent   *DummyNode  `json:"Selector.parent,omitempty"`
	Options  string      `json:"Selector.options,omitempty"`
	Results  []DummyNode `json:"Selector.results,omitempty"`
	DType    []string    `json:"dgraph.type,omitempty"`
}

func (s *Selector) SetDType() {
	s.DType = []string{DType}
}

func (s *Selector) IsValid() bool {
	if s.Created == "" || s.Modified == "" || s.Type == "" || s.Status == "" || s.Options == "" {
		return false
	}

	return validTypes[s.Type] && validStates[s.Status]
}

type AmountRange struct {
	Min *int64
	Max *int64
}

func (a AmountRange) IsValid() bool {
	return !((a.Min == nil && a.Max == nil) || (a.Max != nil && a.Min != nil && *a.Min > *a.Max))
}

type Options struct {
	StartDate                  *time.Time   `json:"startDate,omitempty"`
	EndDate                    *time.Time   `json:"endDate,omitempty"`
	PrivacyTypes               []int        `json:"privacyTypes,omitempty"`
	ExcludePrivacyTransactions *bool        `json:"excludePrivacyTransactions,omitempty"`
	InputSum                   *AmountRange `json:"inputSum,omitempty"`
	OutputSum                  *AmountRange `json:"outputSum,omitempty"`
	InputRange                 *AmountRange `json:"inputRange,omitempty"`
	OutputRange                *AmountRange `json:"outputRange,omitempty"`
}

func (o Options) isValid() bool {
	// both dates must be set
	if o.StartDate == nil || o.EndDate == nil {
		return false
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
		if privacyType < 0 || privacyType > 4 {
			return false
		}
	}

	return true
}

type FrontendSelector struct {
	UID      string `json:"uid,omitempty"`
	Created  string `json:"created,omitempty"`
	Modified string `json:"modified,omitempty"`
	Type     string `json:"type,omitempty"`
	Status   string `json:"status,omitempty"`
	Options  string `json:"options,omitempty"`
	Results  []struct {
		Hash string `json:"txhash,omitempty"`
	} `json:"results,omitempty"`
}
