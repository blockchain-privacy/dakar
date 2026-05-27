// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

var ErrInvalidSelector = errors.New("invalid selector")

// appendFilterArgs appends '<and> ge(filterSubject, number)' or '<and> le(filterSubject, number)' filter
// to the given string and returns it
func appendFilterArgs(filter string, filterSubject string, number *int64, greaterOrEqual bool) string {
	if number == nil {
		return filter
	}

	if filter != "" {
		filter += " and "
	}

	if greaterOrEqual {
		filter += "ge"
	} else {
		filter += "le"
	}

	return filter + "(" + filterSubject + "," + strconv.FormatInt(*number, 10) + ")"
}

// DoSelection returns transactions specified by the options. It also returns the number of total results.
func DoSelection(ctx context.Context, c external.Database, o TxPropOptions, parentUID string) ([]string, int, error) {
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

	var rangeFilter string
	if inputRangeFilter != "" || outputRangeFilter != "" {
		rangeFilter = `@cascade{
					` + inputRangeFilter + `
					` + outputRangeFilter + `
					}`
	}

	var transactionTypeFilter string
	if o.TransactionTypes != nil {
		for _, txType := range o.TransactionTypes {
			if transactionTypeFilter != "" {
				transactionTypeFilter += ","
			}
			transactionTypeFilter += "\"" + txType + "\""
		}

		transactionTypeFilter = "eq(Transaction.type," + transactionTypeFilter + ")"
	}

	// construct @filter(eq(Transaction.type, ..., ...) or eq(Transaction.type, ..., ....) ...)
	if transactionTypeFilter != "" {
		transactionTypeFilter = "@filter(" + transactionTypeFilter + ")"
	}

	if o.ExcludePrivacyTransactions != nil && *o.ExcludePrivacyTransactions {
		transactionTypeFilter = "@filter(not has(Transaction.type))"
	}

	var selectorQuery string

	if parentUID == "" {
		selectorQuery = `var(func: between(ts,"` + o.StartDate.Format(time.RFC3339) + `","` + o.EndDate.Format(time.RFC3339) + `")){
							t as transactions` + transactionTypeFilter + `
						}`
	} else {
		selectorQuery = `
					var(func: uid(` + parentUID + `))@filter(eq(Selector.type, ` + constants.TypeHeuristic + `)){
						Selector.results{
							hr as HeuristicCluster.results` + transactionTypeFilter + `
						}
					}

					var(func: uid(` + parentUID + `))@filter(not eq(Selector.type, ` + constants.TypeHeuristic + `)){
						sr as Selector.results` + transactionTypeFilter + `
					}
					t as var(func: uid(hr,sr))
					`
	}

	maxItems := selectorMaxItems
	if o.MaxItems != nil {
		maxItems = *o.MaxItems
	}

	query := `{
				` + selectorQuery + `

				f as var(func: uid(t))` + rangeFilter + `

				withSums as var(func: uid(f)){
					` + queryBody + `
				}

				count(func: uid(withSums))` + queryFilter + `{
					count:count(uid)
				}
				
				q(func: uid(withSums), first: ` + strconv.Itoa(maxItems) + `)` + queryFilter + `{
					uid
				}
			  }`

	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		return nil, 0, serror.New(err)
	}

	var r struct {
		Count []struct {
			Count int `json:"count,omitempty"`
		} `json:"count,omitempty"`
		Q []db.UIDNode `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, 0, serror.New(err)
	}

	if len(r.Count) != 1 {
		return nil, 0, serror.FromStr("invalid result")
	}

	uids := make([]string, len(r.Q))
	for i, ts := range r.Q {
		uids[i] = ts.UID
	}

	return uids, r.Count[0].Count, nil
}

// InsertSelector inserts the given selector into the database. Returns its UID.
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

	now := time.Now().UTC().Format(time.RFC3339)
	s.Created = now
	s.Modified = now

	pb, err := json.Marshal(dummyWorkspace{UID: workspaceUID, Selectors: []Selector{*s}})
	if err != nil {
		return "", serror.New(err)
	}

	vars := map[string]string{"$userUID": userUID, "$workspaceUID": workspaceUID}

	// for safety the parent UID and workspace UID needs to be checked
	var variable string
	var parentQuery string
	var parentUnion string
	cond := "@if(eq(len(w), 1))"
	if s.Parent != nil {
		vars["$parent"] = s.Parent.UID
		variable = ",$parent: string"
		// parent can either be a heuristic or a selector
		parentQuery = `{
							h as Workspace.selectors@filter(uid($parent))
						}`
		parentUnion = ` t as var(func: uid($parent))@filter(has(txhash))
						p as var(func: uid(t,h))`

		cond = "@if(eq(len(w), 1) and eq(len(p),1))"
	}

	req := &api.Request{
		Query: `query Q($userUID: string, $workspaceUID: string` + variable + `) {
					var(func: uid($userUID))@filter(type(User)){
						w as User.workspaces@filter(uid($workspaceUID))` + parentQuery + `
					}

					` + parentUnion + `
				  }`,
		Vars: vars,
		Mutations: []*api.Mutation{{
			Cond:    cond,
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

// DoGraphSelection returns transactions specified by the options. It also returns the number of total results.
func DoGraphSelection(ctx context.Context, c external.Database, o TxGraphOptions, parentUID string) ([]string, int, error) {
	maxItems := selectorMaxItems
	if o.MaxItems != nil {
		maxItems = *o.MaxItems
	}

	var lookupDirectionQuery string
	var typeFilter string
	if o.ExcludePrivacyTransactions {
		typeFilter = "@filter(not has(Transaction.type))"
	}

	if o.IsForward {
		lookupDirectionQuery = `
								tx_outputs
								t as ~tx_inputs` + typeFilter + `
								`
	} else {
		lookupDirectionQuery = `
								tx_inputs
								t as ~tx_outputs` + typeFilter + `
								`
	}

	query := `query Q($depth:int, $parent:string){
				var(func: uid($parent))@recurse(depth:$depth, loop: false){
					` + lookupDirectionQuery + `
				}
				
				count(func: uid(t)){
					count:count(uid)
				}

				q(func: uid(t), first: ` + strconv.Itoa(maxItems) + `){
					uid
				}
			  }`

	// depth has to be doubled because in recurse both outputs and transactions are traversed
	resp, err := c.Query(ctx, query, map[string]string{"$depth": strconv.Itoa(*o.Depth * 2), "$parent": parentUID})
	if err != nil {
		return nil, 0, serror.New(err)
	}

	var r struct {
		Count []struct {
			Count int `json:"count,omitempty"`
		} `json:"count,omitempty"`
		Q []db.UIDNode `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, 0, serror.New(err)
	}

	if len(r.Count) != 1 {
		return nil, 0, serror.FromStr("invalid result")
	}

	uids := make([]string, len(r.Q))
	for i, ts := range r.Q {
		uids[i] = ts.UID
	}

	return uids, r.Count[0].Count, nil
}

// UpdateSelector updates a selector. Modifying the selector's parent is not allowed.
func UpdateSelector(ctx context.Context, c external.Database, s *Selector, userUID string, workspaceUID string) error {
	if s == nil || s.UID == "" || s.Parent != nil {
		return serror.New(ErrInvalidSelector)
	}

	if s.Type != "" && !IsTypeValid(s.Type) {
		return serror.New(ErrInvalidSelector)
	}

	if s.Status != "" && !IsStatusValid(s.Status) {
		return serror.New(ErrInvalidSelector)
	}

	s.Modified = time.Now().UTC().Format(time.RFC3339)

	pb, err := json.Marshal(*s)
	if err != nil {
		return serror.New(err)
	}

	req := &api.Request{
		Query: `query Q($userUID: string, $workspaceUID: string, $selectorUID: string) {
					var(func: uid($userUID))@filter(type(User)){
						User.workspaces@filter(uid($workspaceUID)) {
							s as Workspace.selectors@filter(uid($selectorUID))
						}
					}
				  }`,
		Vars: map[string]string{"$userUID": userUID, "$workspaceUID": workspaceUID, "$selectorUID": s.UID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(eq(len(s), 1))",
			SetJson: pb,
		}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}

// GetSelectorResultsByUID returns the selector for the given selectorUID, which was created by userUID
func GetSelectorResultsByUID(ctx context.Context, c external.Database,
	selectorUID string, userUID string, workspaceUID string) ([]TransactionWithTimestamp, error) {
	const query = `query Q($selectorUID:string,$userUID:string,$workspaceUID:string){
				var(func: uid($userUID)){
					User.workspaces@filter(uid($workspaceUID)){
						Workspace.selectors@filter(uid($selectorUID)){
							r as Selector.results
						}
					}
				}

				transactions(func: uid(r))@filter(has(txhash))@normalize{
					txhash:txhash
					txtype:Transaction.type
					~transactions{
						ts:ts
					}
				}

				clusters(func: uid(r))@filter(has(HeuristicCluster.results)){
					transactions: HeuristicCluster.results@normalize{
						txhash:txhash
						txtype:Transaction.type
						~transactions{
							ts:ts
						}
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
		// set if not a heuristic
		Transactions []TransactionWithTimestamp `json:"transactions,omitempty"`
		// set if a heuristic
		Clusters []struct {
			Transactions []TransactionWithTimestamp `json:"transactions,omitempty"`
		} `json:"clusters,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	if len(r.Transactions) > 0 {
		return r.Transactions, nil
	}

	if len(r.Clusters) > 0 {
		var allTransactions []TransactionWithTimestamp
		for i, cluster := range r.Clusters {
			for y := range cluster.Transactions {
				cluster.Transactions[y].Cluster = &i
			}

			allTransactions = append(allTransactions, cluster.Transactions...)
		}

		return allTransactions, nil
	}

	return nil, serror.FromStr("no results returned")
}

// DeleteUserSelectors deletes all given selectors of a user
func DeleteUserSelectors(ctx context.Context, c external.Database,
	uids []string, userUID string, workspaceUID string) error {
	const query = `
		query Q($userUID:string,$selectorUIDs:string,$workspaceUID:string){
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					s as Workspace.selectors@filter(uid($selectorUIDs)){
						hc as Selector.results@filter(has(HeuristicCluster.results))
					}
				}
			}
		}`

	req := &api.Request{
		Query: query,
		Vars: map[string]string{"$userUID": userUID,
			"$selectorUIDs": db.CreateCommaArray(uids), "$workspaceUID": workspaceUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hc) * * .
								uid(s) * * .
								<` + workspaceUID + "> <Workspace.selectors> uid(s) ."),
		}},
		CommitNow: true,
	}

	resp, err := c.Mutate(ctx, req)
	if err != nil {
		return serror.New(err)
	}

	if !db.HasMutationCost(resp) {
		return serror.New(db.ErrNoMutationHappened)
	}

	return nil
}

// GetWaitingSelectors returns selectors which are waiting to be executed.
func GetWaitingSelectors(ctx context.Context, c external.Database, maxItems int) ([]WorkItem, error) {
	if maxItems <= 0 {
		return nil, nil
	}

	query := `query Q($maxItems:int){
				q(func: eq(Selector.status, ` + StatusWaiting + `), first: $maxItems){
					uid
					Selector.type
					Selector.options
					Selector.parent{uid}
					~Workspace.selectors{
						uid
						~User.workspaces{uid}
					}
				}
			   }`

	resp, err := c.Query(ctx, query, map[string]string{"$maxItems": strconv.Itoa(maxItems)})
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Selectors []struct {
			UID       string      `json:"uid,omitempty"`
			Type      string      `json:"Selector.type,omitempty"`
			Options   string      `json:"Selector.options,omitempty"`
			Parent    *db.UIDNode `json:"Selector.parent,omitempty"`
			Workspace []struct {
				UID  string       `json:"uid,omitempty"`
				User []db.UIDNode `json:"~User.workspaces,omitempty"`
			} `json:"~Workspace.selectors,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	items := make([]WorkItem, len(r.Selectors))
	for i, s := range r.Selectors {
		if len(s.Workspace) != 1 || len(s.Workspace[0].User) != 1 {
			return nil, serror.FromStr("invalid workspace or user UID")
		}

		var parentUID string
		if s.Parent != nil {
			parentUID = s.Parent.UID
		}

		items[i] = WorkItem{
			UserUID:         s.Workspace[0].User[0].UID,
			WorkspaceUID:    s.Workspace[0].UID,
			ParentUID:       parentUID,
			SelectorUID:     s.UID,
			SelectorType:    s.Type,
			SelectorOptions: s.Options,
		}
	}

	return items, nil
}

// GetSelectorStatus returns the status of the given selector. If the selector does not exist, an error is returned.
func GetSelectorStatus(ctx context.Context, c external.Database, selectorUID string,
	workspaceUID string, userUID string) (string, error) {
	const query = `
		query Q($userUID:string,$selectorUID:string,$workspaceUID:string){
			var(func: uid($userUID)){
				User.workspaces@filter(uid($workspaceUID)){
					s as Workspace.selectors@filter(uid($selectorUID))
				}
			}

			q(func: uid(s)){
				Selector.status
			}
		}`

	resp, err := c.Query(ctx, query, map[string]string{"$selectorUID": selectorUID, "$workspaceUID": workspaceUID, "$userUID": userUID})
	if err != nil {
		return "", serror.New(err)
	}

	var r struct {
		Selectors []struct {
			Status string `json:"Selector.status,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", serror.New(err)
	}

	if len(r.Selectors) != 1 {
		return "", serror.New(ErrInvalidSelector)
	}

	return r.Selectors[0].Status, nil
}
