package transaction

import (
	"backend/cmd/cliutil"
	dbtxh "backend/db/analytics/heuristics/transaction"
	"backend/external"

	"errors"
	"fmt"
	"strings"
)

// validHeuristicTypes includes all heuristics which are possible to receive from the frontend.
// New heuristics must be added here
var validHeuristicTypes = []heuristic{NewOneSourceHeuristic(0), NewAmountHeuristic(),
	NewPerfectMatchHeuristic(), NewDenominationTypeHeuristic(), NewTimeConstraintHeuristic(0),
	NewForwardTimeHeuristic(0)}

// typeMap K: heuristic types, v: heuristics
var typeMap = make(map[string]heuristic)

// newUIDPrefix is the prefix of the uid of all newly created heuristics on the frontend
const newUIDPrefix = "newUid_"

// errors for this file
var (
	errHeuristicTypeNotFound     = errors.New("error heuristic type not found")
	errHeuristicDuplicateUID     = errors.New("error duplicate uids found")
	errHeuristicUIDNotFound      = errors.New("error uid not found")
	errHeuristicMultipleRoots    = errors.New("error multiple roots found")
	errHeuristicNoRoots          = errors.New("error no roots") // ♩ ♬ no roots ♬ ♪
	errHeuristicNotValid         = errors.New("error heuristics are not valid")
	errHeuristicNoExecutorsBuilt = errors.New("error no executors have been built")
)

type heuristicTreeElement struct {
	uid                string
	parentHeuristicUID string
	heuristic          heuristic
	childHeuristicUID  []string
}

// isValid checks if the given heuristics are all valid
func isValid(hMap map[string]heuristic, heuristics []dbtxh.FrontendHeuristic) bool {
	if len(heuristics) == 0 {
		return false
	}

	for _, h := range heuristics {
		// more than one parent is not allowed
		if len(h.ParentHeuristic) > 1 {
			return false
		}

		// type must by in valid set; parameter must be set if the map has a parameter
		if modelHeuristic, ok := hMap[h.Type]; !ok || (modelHeuristic.hasParameter() && len(h.Parameter) == 0) {
			return false
		}
	}

	return true
}

func buildHeuristicTreeElements(hMap map[string]heuristic, heuristics []dbtxh.FrontendHeuristic) (builtHeuristics map[string]heuristicTreeElement,
	err error) {
	// create map
	builtHeuristics = make(map[string]heuristicTreeElement)

	// add elements to map
	for _, h := range heuristics {
		// create new heuristic

		if modelHeuristic, ok := hMap[h.Type]; ok {
			newHeuristic := modelHeuristic.clone()

			//newHeuristic := *modelHeuristic
			// check heuristic was already built
			if _, ok := builtHeuristics[h.UID]; ok {
				err = errHeuristicDuplicateUID
				return
			}

			if newHeuristic.hasParameter() {
				err = newHeuristic.setParameter(h.Parameter)
				if err != nil {
					err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
					return
				}
			}

			var childHeuristicUids []string

			for _, c := range h.ChildHeuristics {
				childHeuristicUids = append(childHeuristicUids, c.UID)
			}

			var parentUID string

			if len(h.ParentHeuristic) > 0 {
				parentUID = h.ParentHeuristic[0].UID
			}

			builtHeuristics[h.UID] = heuristicTreeElement{
				uid:                h.UID,
				parentHeuristicUID: parentUID,
				heuristic:          newHeuristic,
				childHeuristicUID:  childHeuristicUids,
			}

		} else {
			err = errHeuristicTypeNotFound
			return
		}
	}

	return
}

type treeLeaf struct {
	uid   string
	level int
}

// traverseHeuristicTree is a recursive function called by getNodeLevelDistribution. It traverses a tree starting
// at root and returns the node level of itself and all child nodes
func traverseHeuristicTree(nodes map[string]heuristicTreeElement, rootUID string, root heuristicTreeElement,
	level int) (l []treeLeaf, err error) {
	for _, uid := range root.childHeuristicUID {
		if h, ok := nodes[uid]; ok {

			leafs, traverseErr := traverseHeuristicTree(nodes, uid, h, level+1)
			if traverseErr != nil {
				err = traverseErr
				return
			}

			l = append(l, leafs...)
		} else {
			err = errHeuristicUIDNotFound
			return
		}
	}

	l = append(l, treeLeaf{
		uid:   rootUID,
		level: level,
	})

	return
}

// getNodeLevelDistribution returns a list of all nodes with the respective level in the tree
func getNodeLevelDistribution(nodes map[string]heuristicTreeElement, rootUID string,
	root heuristicTreeElement) (levelToNode [][]heuristicTreeElement, err error) {
	treeNodes, err := traverseHeuristicTree(nodes, rootUID, root, 0)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, n := range treeNodes {
		level := n.level
		twoDimLength := len(levelToNode)
		// add elements to first dimension of array if it not exist
		if twoDimLength < level+1 {
			for i := 0; i < level+1-twoDimLength; i++ {
				levelToNode = append(levelToNode, []heuristicTreeElement{})
			}
		}

		// don't worry, the slice is not nil. It is always initialised in the above if clause
		levelToNode[level] = append(levelToNode[level], nodes[n.uid])
	}

	return
}

// buildExecutorsFromLevels builds a HeuristicExecutor beginning at the deepest level of the tree
func buildExecutorsFromLevels(levelToNode [][]heuristicTreeElement) (rootExecutor HeuristicExecutor, err error) {
	executorStack := make(map[string]HeuristicExecutor)

	// backward traversal of list, because the deepest level has the highest index
	for i := len(levelToNode) - 1; i >= 0; i-- {
		for _, n := range levelToNode[i] {
			thisExec := BuildExecutor(n.heuristic)
			thisExec.RootUID = n.parentHeuristicUID
			for _, childUID := range n.childHeuristicUID {
				if child, ok := executorStack[childUID]; ok {
					// add child to this executor tree and delete it from the map
					thisExec.NextHeuristics = append(thisExec.NextHeuristics, child)
					delete(executorStack, childUID)
				} else {
					err = errHeuristicUIDNotFound
					return
				}
			}

			// add new executor to map
			executorStack[n.uid] = thisExec
		}
	}

	// remaining element in executorStack must be one -> only one root allowed
	if len(executorStack) != 1 {
		err = errHeuristicMultipleRoots
		return
	}

	// set rootExecutor to the only remaining value
	for _, v := range executorStack {
		rootExecutor = v
	}

	return
}

func isRootHeuristic(h heuristicTreeElement, heuristics map[string]heuristicTreeElement) bool {
	// no parent uid -> must be a root element
	if h.parentHeuristicUID == "" {
		return true
	}

	// parent uid does not exist in change set -> must be a root element in this context
	if _, ok := heuristics[h.parentHeuristicUID]; !ok {
		return true
	}

	return false
}

// buildExecutors builds HeuristicExecutor trees from heuristics starting at the given rootHeuristicUids.
// Each element of the returned slice is one HeuristicExecutor tree.
func buildExecutors(rootHeuristicUids []string, heuristics map[string]heuristicTreeElement) (
	executors []HeuristicExecutor, err error) {
	for _, uid := range rootHeuristicUids {
		v, ok := heuristics[uid]
		if !ok {
			err = errHeuristicUIDNotFound
			return
		}

		if len(v.childHeuristicUID) == 0 {
			executors = append(executors, HeuristicExecutor{
				ThisHeuristic: v.heuristic,
				RootUID:       v.parentHeuristicUID,
			})
			// no children -> no work
			continue
		}

		levelDistribution, childErr := getNodeLevelDistribution(heuristics, v.uid, v)
		if childErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), childErr)
			return
		}

		newRootExecutor, buildErr := buildExecutorsFromLevels(levelDistribution)
		if buildErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), buildErr)
			return
		}

		executors = append(executors, newRootExecutor)
	}

	return
}

// ConstructExecutors creates executors based on heuristics
func ConstructExecutors(dgraph external.Database, txhash string, heuristics []dbtxh.FrontendHeuristic) (
	executors []HeuristicExecutor, err error) {
	// only set values for global type map once
	if len(typeMap) == 0 {
		for _, h := range validHeuristicTypes {
			typeMap[h.getType()] = h
		}
	}

	if !isValid(typeMap, heuristics) {
		err = errHeuristicNotValid
		return
	}

	newHeuristics, err := buildHeuristicTreeElements(typeMap, heuristics)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// collect root uids
	var rootUids []string
	var rootsToCheck []string
	for _, v := range newHeuristics {
		if isRootHeuristic(v, newHeuristics) {
			rootUids = append(rootUids, v.uid)
			// no need to check roots which are newly created and thus have no valid uid
			if !strings.HasPrefix(v.uid, newUIDPrefix) {
				rootsToCheck = append(rootsToCheck, v.uid)
			}
		}
	}

	if len(rootUids) == 0 {
		err = errHeuristicNoRoots
		return
	}

	if len(rootsToCheck) > 0 {
		// check if all parent heuristics of contextual roots actually exists in the db
		if exists, checkErr := dbtxh.DoesHeuristicUIDExist(dgraph, txhash, rootsToCheck); checkErr != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), checkErr)
			return
		} else if !exists {
			err = errHeuristicUIDNotFound
			return
		}
	}

	executors, err = buildExecutors(rootUids, newHeuristics)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// areSetsValid checks if the uids of removed are appearing in changed
func areSetsValid(changed []dbtxh.FrontendHeuristic, removed []string) bool {
	// if both are empty there is an error
	if len(changed) == 0 && len(removed) == 0 {
		return false
	}

	// if one is empty we have nothing to check against
	if len(changed) == 0 || len(removed) == 0 {
		return true
	}

	if len(changed) > 0 {
		// copy uids into map
		changeMap := make(map[string]bool)
		for _, c := range changed {
			changeMap[c.UID] = true
		}

		for _, r := range removed {
			if ok := changeMap[r]; ok {
				return false
			}
		}
	}

	return true
}

// mergeRemoveList adds the uids of
func mergeRemoveList(changed []dbtxh.FrontendHeuristic, removed []string) []string {
	for _, c := range changed {
		// only add heuristics which actually exist in the database
		if !strings.HasPrefix(c.UID, newUIDPrefix) {
			removed = append(removed, c.UID)
		}
	}

	return removed
}

// CreateWork does some checks changed and toRemove
func CreateWork(dgraph external.Database, transactionHash string, changed []dbtxh.FrontendHeuristic, toRemove []string) (w Work, err error) {
	if !areSetsValid(changed, toRemove) {
		err = errors.New("error sets are not valid")
		return
	}

	if len(changed) > 0 {
		// create HeuristicExecutor trees
		w.executors, err = ConstructExecutors(dgraph, transactionHash, changed)
		if err != nil {
			err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
			return
		} else if len(w.executors) == 0 {
			err = errHeuristicNoExecutorsBuilt
			return
		}
	}

	w.removableHeuristics = mergeRemoveList(changed, toRemove)

	return
}
