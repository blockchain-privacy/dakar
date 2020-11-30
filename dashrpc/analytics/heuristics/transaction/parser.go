package transaction

import (
	"dashrpc/cmd/cliutil"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	"errors"
	"fmt"
)

// validHeuristics includes all heuristics which are possible to receive from the frontend.
// New heuristics must be added here
var validHeuristics = []heuristic{NewOneSourceHeuristic(0), NewAmountHeuristic(),
	NewPerfectMatchHeuristic(), NewDenominationTypeHeuristic()}

// errors for this file
var (
	errHeuristicTypeNotFound  = errors.New("error heuristic type not found")
	errHeuristicDuplicateUid  = errors.New("error duplicate uids found")
	errHeuristicUidNotFound   = errors.New("error uid not found")
	errHeuristicMultipleRoots = errors.New("error multiple roots found")
	errHeuristicNotValid      = errors.New("error heuristics are not valid")
)

type heuristicTreeElement struct {
	uid                string
	parentHeuristicUid string
	heuristic          heuristic
	childHeuristicUid  []string
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
		if heuristic, ok := hMap[h.Type]; !ok || (heuristic.hasParameter() && len(h.Parameter) == 0) {
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
		if newHeuristic, ok := hMap[h.Type]; ok {
			if _, ok := builtHeuristics[h.Uid]; ok {
				err = errHeuristicDuplicateUid
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
				childHeuristicUids = append(childHeuristicUids, c.Uid)
			}

			var parentUid string

			if len(h.ParentHeuristic) > 0 {
				parentUid = h.ParentHeuristic[0].Uid
			}

			builtHeuristics[h.Uid] = heuristicTreeElement{
				uid:                h.Uid,
				parentHeuristicUid: parentUid,
				heuristic:          newHeuristic,
				childHeuristicUid:  childHeuristicUids,
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
func traverseHeuristicTree(nodes map[string]heuristicTreeElement, rootUid string, root heuristicTreeElement,
	level int) (l []treeLeaf, err error) {
	for _, uid := range root.childHeuristicUid {
		if h, ok := nodes[uid]; ok {

			leafs, traverseErr := traverseHeuristicTree(nodes, uid, h, level+1)
			if traverseErr != nil {
				err = traverseErr
				return
			}

			l = append(l, leafs...)
		} else {
			err = errHeuristicUidNotFound
			return
		}
	}

	l = append(l, treeLeaf{
		uid:   rootUid,
		level: level,
	})

	return
}

// getNodeLevelDistribution returns a list of all nodes with the respective level in the tree
func getNodeLevelDistribution(nodes map[string]heuristicTreeElement, rootUid string,
	root heuristicTreeElement) (levelToNode [][]heuristicTreeElement, err error) {
	treeNodes, err := traverseHeuristicTree(nodes, rootUid, root, 0)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, n := range treeNodes {
		level := n.level
		twoDimLength := len(levelToNode)
		// add elements to first dimension of array if not exist
		if twoDimLength < level+1 {
			for i := 0; i < level+1-twoDimLength; i++ {
				levelToNode = append(levelToNode, []heuristicTreeElement{})
			}
		}

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

			for _, childUid := range n.childHeuristicUid {
				if child, ok := executorStack[childUid]; ok {
					// add child to this executor tree and delete it from the map
					thisExec.NextHeuristics = append(thisExec.NextHeuristics, child)
					delete(executorStack, childUid)
				} else {
					err = errHeuristicUidNotFound
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

func buildExecutors(heuristics map[string]heuristicTreeElement) (executors []HeuristicExecutor, err error) {
	//processedElements := make(map[string]heuristicTreeElement)

	var roots []heuristicTreeElement
	for k, v := range heuristics {
		// start with root nodes
		if v.parentHeuristicUid == "" {
			roots = append(roots, v)

			if len(v.childHeuristicUid) == 0 {
				executors = append(executors, HeuristicExecutor{
					ThisHeuristic: v.heuristic,
				})
				// no children -> no work
				continue
			}

			levelDistribution, childErr := getNodeLevelDistribution(heuristics, k, v)
			if childErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), childErr)
				return
			}

			if newRootExecutor, buildErr := buildExecutorsFromLevels(levelDistribution); buildErr != nil {
				err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), buildErr)
				return
			} else {
				executors = append(executors, newRootExecutor)
			}
		}
	}

	return
}

func ConstructExecutors(heuristics []dbtxh.FrontendHeuristic) (executors []HeuristicExecutor, err error) {
	heuristicMap := make(map[string]heuristic)

	for _, h := range validHeuristics {
		heuristicMap[h.getType()] = h
	}

	if !isValid(heuristicMap, heuristics) {
		err = errHeuristicNotValid
		return
	}

	newHeuristics, err := buildHeuristicTreeElements(heuristicMap, heuristics)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	executors, err = buildExecutors(newHeuristics)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}
