package selectors

import (
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
	StartDate   *time.Time
	EndDate     *time.Time
	OutputSum   *AmountRange
	InputSum    *AmountRange
	InputRange  *AmountRange
	OutputRange *AmountRange
}

func (o Options) isValid() bool {
	// both dates must be set
	if o.StartDate == nil || o.EndDate == nil {
		return false
	}

	// at least one option must be set
	if o.OutputSum == nil && o.InputSum == nil && o.InputRange == nil && o.OutputRange == nil {
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

	query := `{
				var(func: between(ts,"` + o.StartDate.Format(time.RFC3339) + `","` + o.EndDate.Format(time.RFC3339) + `")){
					t as transactions
				}

				filterByOutputs as var(func: uid(t))@cascade{
					` + inputRangeFilter + `
					` + outputRangeFilter + `
				}

				withSums as var(func: uid(filterByOutputs)){
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
