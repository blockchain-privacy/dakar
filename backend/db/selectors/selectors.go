package selectors

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"context"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

var ErrInvalidOptions = errors.New("invalid options")
var ErrInvalidSelector = errors.New("invalid selector")

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

// InsertSelector inserts the given selector into the database
func InsertSelector(ctx context.Context, c external.Database, s *Selector,
	userUID string, workspaceUID string) (string, error) {
	if s == nil || !s.IsValid() {
		return "", serror.New(ErrInvalidSelector)
	}

	s.SetDType()

	const newSelectorDummyUID = "new_h"
	s.UID = "_:" + newSelectorDummyUID

	type dummyWorkspace struct {
		UID       string     `json:"uid,omitempty"`
		Selectors []Selector `json:"Workspace.selectors,omitempty"`
	}

	pb, err := json.Marshal(dummyWorkspace{UID: workspaceUID, Selectors: []Selector{*s}})
	if err != nil {
		return "", serror.New(err)
	}

	req := &api.Request{
		Query: `query Q($userUID: string, $workspaceUID: string) {
					var(func: uid($userUID))@filter(type(User)){
						w as User.workspaces@filter(uid($workspaceUID))
					}
				  }`,
		Vars: map[string]string{"$userUID": userUID, "$workspaceUID": workspaceUID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(gt(len(w), 0))",
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return "", err
	}

	insertUID, ok := resp.GetUids()[newSelectorDummyUID]
	if !ok {
		return "", serror.FromStr("no new selector created")
	}

	return insertUID, nil
}

// GetFrontendSelectorByUID returns the selector for the given selectorUID, which was created by userUID
func GetFrontendSelectorByUID(ctx context.Context, c external.Database,
	selectorUID string, userUID string, workspaceUID string) (*FrontendSelector, error) {
	const query = `query Q($selectorUID:string,$userUID:string,$workspaceUID:string){
				var(func: uid($userUID)){
					User.workspaces@filter(uid($workspaceUID)){
						s as Workspace.selectors@filter(uid($selectorUID))
					}
				}

				q(func: uid(s)){
					created: Selector.created
					modified: Selector.modified
					type: Selector.type
					status: Selector.status
					options: Selector.options
					results: Selector.results {
						txhash
					}
				}
			   }`

	resp, err := c.Query(ctx, query, map[string]string{"$selectorUID": selectorUID,
		"$userUID": userUID, "$workspaceUID": workspaceUID})
	if err != nil {
		return nil, serror.New(err)
	}

	// json struct
	var r struct {
		Selectors []FrontendSelector `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.New(err)
	}

	if len(r.Selectors) == 0 {
		return nil, serror.FromStr("no selector returned")
	}

	r.Selectors[0].UID = selectorUID

	return &r.Selectors[0], nil
}
