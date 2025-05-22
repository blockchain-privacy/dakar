package graph

import (
	"backend/constants"
	"github.com/qrest/gomisc/serror"
	"math"
	"sort"
)

func getEarliestTimestamp(g *ReversibleGraph, tx TransactionNode) (int64, error) {
	rootInputs := g.From(tx.ID())
	if rootInputs.Len() == 0 {
		return 0, serror.FromStrWithContext("transaction has not timestamps", "tx", tx)
	}

	var lowestTimestamp int64
	for rootInputs.Next() {
		node, ok := rootInputs.Node().(TransactionNode)
		if !ok {
			continue
		}

		ts := node.TS.Unix()
		if lowestTimestamp == 0 || ts < lowestTimestamp {
			lowestTimestamp = ts
		}
	}
	return lowestTimestamp, nil
}

type FingerPrint struct {
	TransactionUID string
	Score          float64
	SessionCount   int
}

// SpendingFingerprint returns a list of transaction uids which have a similar spending pattern
// and the number of mixing sessions of this transactions. Uses the Chamfer distance as a similarity measure.
func SpendingFingerprint(g *ReversibleGraph, uid string) ([]FingerPrint, int, error) {
	// maximumDistance is the maximum distance between to earliest (lowest) input timestamp
	// of the root transaction and the timestamp of the compared transaction
	// 2 days = 60 * 60 * 24 * 2 = 172800 seconds
	const maximumDistance = 172800

	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, -1, err
	}

	rootNode := g.Node(nodeUID)
	if rootNode == nil {
		return nil, -1, serror.FromStr(uid + " not in graph")
	}

	rootTx, ok := rootNode.(TransactionNode)
	if !ok || !constants.IsDestinationTransaction(rootTx.Type) {
		return nil, -1, serror.FromStr(uid + " is not a destination transaction")
	}

	earliestInputTimestamp, err := getEarliestTimestamp(g, rootTx)
	if err != nil {
		return nil, -1, err
	}

	const maxNumberOfScoreResults = 30
	var fingerprints []FingerPrint
	nodes := g.Nodes()
	for nodes.Next() {
		node := nodes.Node()

		// do not compare with itself
		if node.ID() == nodeUID {
			continue
		}

		txNode, ok := node.(TransactionNode)
		// filter for exact transaction type, so we don't match transactions from different coinjoin types
		if !ok || txNode.Type != rootTx.Type || earliestInputTimestamp-txNode.TS.Unix() > maximumDistance {
			continue
		}

		dist := calcChamferDistance(g, rootTx, txNode)
		// only consider average distances of less or equal than 24h
		if dist > 86400 {
			continue
		}

		fingerprints = append(fingerprints, FingerPrint{
			TransactionUID: txNode.String(),
			Score:          dist,
		})
		sort.Slice(fingerprints, func(i, j int) bool {
			return fingerprints[i].Score > fingerprints[j].Score
		})

		// remove the first element (which has the lowest score)
		if len(fingerprints) > maxNumberOfScoreResults {
			fingerprints = fingerprints[1:]
		}
	}

	return fingerprints, getSessionCount(g, rootTx), err
}

func getFloatTimestamps(g *ReversibleGraph, node TransactionNode) []float64 {
	rootInputs := g.From(node.ID())
	if rootInputs.Len() == 0 {
		return nil
	}

	var timestamps []float64
	for rootInputs.Next() {
		node, ok := rootInputs.Node().(TransactionNode)
		if !ok {
			continue
		}

		timestamps = append(timestamps, float64(node.TS.Unix()))
	}

	return timestamps
}

func calcChamferDistance(g *ReversibleGraph, node1 TransactionNode, node2 TransactionNode) float64 {
	node1Timestamps := getFloatTimestamps(g, node1)
	node2Timestamps := getFloatTimestamps(g, node2)

	return chamferDistanceOneSided(node1Timestamps, node2Timestamps)
}

// chamferDistanceOneSided calculates the one sided Chamfer distance between two 1D arrays.
func chamferDistanceOneSided(arr1, arr2 []float64) float64 {
	if len(arr1) == 0 || len(arr2) == 0 {
		return math.Inf(1) // Return infinity if either array is empty
	}

	totalDistance := 0.0
	for _, a1 := range arr1 {
		minDistance := math.Inf(1) // Start with infinity
		for _, a2 := range arr2 {
			distance := math.Abs(a1 - a2)
			if distance < minDistance {
				minDistance = distance
			}
		}
		totalDistance += minDistance
	}

	return totalDistance / float64(len(arr1))
}

// getSessionCount splits the input timestamps into groups and returns for each group its mean
func getSessionCount(g *ReversibleGraph, tx TransactionNode) int {
	rootInputs := g.From(tx.ID())
	if rootInputs.Len() == 0 {
		return 0
	}

	var timestamps []int64
	for rootInputs.Next() {
		node, ok := rootInputs.Node().(TransactionNode)
		if !ok {
			continue
		}
		timestamps = append(timestamps, node.TS.Unix())
	}

	return len(splitTimestampsIntoSessions(timestamps))
}

// splitTimestampsIntoSessions splits the given timestamps into groups
// based on the duration between them. The groups are ordered increasing by time.
func splitTimestampsIntoSessions(timestamps []int64) [][]int64 {
	if len(timestamps) == 0 {
		return nil
	}

	// sort timestamps
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] < timestamps[j]
	})

	// number of seconds in 12 hours
	const splitDuration = 60 * 60 * 12

	var sessions [][]int64
	var session []int64 //nolint:prealloc
	for i, t := range timestamps {
		session = append(session, t)

		if i+1 == len(timestamps) {
			break
		}

		if timestamps[i+1]-t > splitDuration {
			sessions = append(sessions, session)
			session = []int64{}
		}
	}

	if len(session) > 0 {
		sessions = append(sessions, session)
	}

	return sessions
}
