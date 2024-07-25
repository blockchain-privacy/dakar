package selectors

import (
	"backend/constants"
	"backend/external"
	"context"
	"encoding/json"
	"errors"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

var ErrInvalidOptions = errors.New("invalid options")

type AmountRange struct {
	Min *int64
	Max *int64
}

func (a AmountRange) IsValid() bool {
	return !((a.Min == nil && a.Max == nil) || (a.Max != nil && a.Min != nil && *a.Min > *a.Max))
}

type Options struct {
	StartDate                  *time.Time
	EndDate                    *time.Time
	PrivacyTypes               []int
	ExcludePrivacyTransactions *bool
	InputSum                   *AmountRange
	OutputSum                  *AmountRange
	InputRange                 *AmountRange
	OutputRange                *AmountRange
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

// appendFilterArgs appends '<and> ge(filterSubject, number)' or '<and> le(filterSubject, number)' filter
// to the given string and returns it
func appendFilterArgs(filter string, fiterSubject string, number *int64, min bool) string {
	if number == nil {
		return filter
	}

	if filter != "" {
		filter += " and "
	}

	if min {
		filter += "ge"
	} else {
		filter += "le"
	}

	return filter + "(" + fiterSubject + "," + strconv.FormatInt(*number, 10) + ")"
}

func DoSelection(ctx context.Context, c external.Database, o Options) ([]string, error) {
	if !o.isValid() {
		return nil, serror.New(ErrInvalidOptions)
	}

	var queryBody string
	var queryFilter string

	if o.InputSum != nil {
		queryBody += `
					tx_inputs{
						inputval as amount
					}
					inputsum as sum(val(inputval))
					`

		queryFilter = appendFilterArgs(queryFilter, "val(inputsum)", o.InputSum.Min, true)
		queryFilter = appendFilterArgs(queryFilter, "val(inputsum)", o.InputSum.Max, false)
	}

	if o.OutputSum != nil {
		queryBody += `
					tx_outputs{
						outputval as amount
					}
					outputsum as sum(val(outputval))
					`

		queryFilter = appendFilterArgs(queryFilter, "val(outputsum)", o.OutputSum.Min, true)
		queryFilter = appendFilterArgs(queryFilter, "val(outputsum)", o.OutputSum.Max, false)
	}

	// construct @filter(ge(amount, ...) and le(amount, ...))
	if queryFilter != "" {
		queryFilter = "@filter(" + queryFilter + ")"
	}

	var inputRangeFilter string
	if o.InputRange != nil {
		inputRangeFilter = appendFilterArgs(inputRangeFilter, "amount", o.InputRange.Min, true)
		inputRangeFilter = appendFilterArgs(inputRangeFilter, "amount", o.InputRange.Max, false)
	}

	// construct tx_inputs@filter(ge(amount, ...) and le(amount, ...)){amount}
	if inputRangeFilter != "" {
		inputRangeFilter = "tx_inputs@filter(" + inputRangeFilter + "){amount}"
	}

	var outputRangeFilter string
	if o.OutputRange != nil {
		outputRangeFilter = appendFilterArgs(outputRangeFilter, "amount", o.OutputRange.Min, true)
		outputRangeFilter = appendFilterArgs(outputRangeFilter, "amount", o.OutputRange.Max, false)
	}

	// construct tx_outputs@filter(ge(amount, ...) and le(amount, ...)){amount}
	if outputRangeFilter != "" {
		outputRangeFilter = "tx_outputs@filter(" + outputRangeFilter + "){amount}"
	}

	var privacyTypeFilter string
	if o.PrivacyTypes != nil {
		for _, privacyType := range o.PrivacyTypes {
			if privacyTypeFilter != "" {
				privacyTypeFilter += " or "
			}

			switch privacyType {
			case 0:
				privacyTypeFilter += "between(privacytype,0," + constants.StrPrivacyMixingLast + ")"
			case 1:
				privacyTypeFilter += "between(privacytype," + constants.StrPrivacyDestinationFirst + "," + constants.StrPrivacyDestinationLast + ")"
			case 2:
				privacyTypeFilter += "between(privacytype," + constants.StrPrivacyOriginFirst + "," + constants.StrPrivacyOriginLast + ")"
			case 3:
				privacyTypeFilter += "between(privacytype," + constants.StrPrivacyCollateralCreationFirst + "," + constants.StrPrivacyCollateralCreationLast + ")"
			case 4:
				privacyTypeFilter += "between(privacytype," + constants.StrPrivacyCollateralPaymentFirst + "," + constants.StrPrivacyCollateralPaymentLast + ")"
			}
		}
	}

	// construct @filter(between(privacytype, ..., ...) or between(privacytype, ..., ....) ...)
	if privacyTypeFilter != "" {
		privacyTypeFilter = "@filter(" + privacyTypeFilter + ")"
	}

	if o.ExcludePrivacyTransactions != nil && *o.ExcludePrivacyTransactions {
		privacyTypeFilter = "@filter(not has(privacytype))"
	}

	query := `{
				var(func: between(ts,"` + o.StartDate.Format(time.RFC3339) + `","` + o.EndDate.Format(time.RFC3339) + `")){
					t as transactions` + privacyTypeFilter + `@cascade{
					` + inputRangeFilter + `
					` + outputRangeFilter + `
					}
				}

				withSums as var(func: uid(t)){
					` + queryBody + `
				}
				
				q(func: uid(withSums))` + queryFilter + `{
					uid
				}
			  }`

	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Q []struct {
			UID string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.New(err)
	}

	uids := make([]string, len(r.Q))
	for i, ts := range r.Q {
		uids[i] = ts.UID
	}

	return uids, nil
}
