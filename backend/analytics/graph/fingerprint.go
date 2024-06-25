package graph

import (
	"github.com/qrest/gomisc/serror"
	"math"
	"sort"
)

// getSessionsFromTransaction splits the input timestamps into groups and returns for each group its mean
func getSessionsFromTransaction(g *ReversibleGraph, tx TransactionNode) []int64 {
	rootInputs := g.From(tx.ID())
	if rootInputs.Len() == 0 {
		return nil
	}

	timestampMap := map[int64]bool{}

	for rootInputs.Next() {
		node, ok := rootInputs.Node().(TransactionNode)
		if !ok {
			continue
		}
		timestampMap[node.TS.Unix()] = true
	}

	timestamps := make([]int64, len(timestampMap))
	i := 0
	for k := range timestampMap {
		timestamps[i] = k
		i++
	}

	return getSessionMeans(splitTimestampsIntoSessions(timestamps))
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

// getSessionMeans returns for each session the mean
func getSessionMeans(sessions [][]int64) []int64 {
	means := make([]int64, len(sessions))
	for i, session := range sessions {
		var sessionMean int64

		for _, timestamp := range session {
			sessionMean += timestamp
		}
		// don't care about float cutoff, 1 second more or less does not matter
		means[i] = sessionMean / int64(len(session))
	}
	return means
}

// getShortestDistances returns all distances for each closest sessions pair
func getShortestDistances(sessionMeans1 []int64, sessionMeans2 []int64) []int64 {
	smallestDistances := make([]int64, len(sessionMeans1))

	// for each mean in sessionsMeans1 find the mean in sessionMeans2 which has the shortest distance
	for i, mean1 := range sessionMeans1 {
		var smallestDistance = int64(-1)

		for _, mean2 := range sessionMeans2 {
			distance := mean1 - mean2
			if distance < 0 {
				distance *= -1
			}

			if smallestDistance == int64(-1) || distance < smallestDistance {
				smallestDistance = distance
			}
		}
		smallestDistances[i] = smallestDistance
	}

	return smallestDistances
}

// scoreMeans calculates the score based on the distances between the means
func scoreMeans(rootMeans []int64, otherMeans []int64) (float64, error) {
	distances := getShortestDistances(rootMeans, otherMeans)
	if len(distances) == 0 {
		return -1, serror.NewStackErrorStr("no distances found")
	}

	var score float64

	for _, d := range distances {
		// seconds to hours
		distance := float64(d) / 60 / 60
		if distance == 0 {
			distance = 0.000001
		}

		score += 1 / (math.Pow(distance, float64(1)/float64(4)) + 0.01*math.Pow(distance, 3))
	}

	return score, nil
}

type FingerPrint struct {
	TransactionUID string
	Score          float64
	SessionCount   int
}

// SpendingFingerprint returns a list of transaction uids which have a similar spending pattern
// and the number of mixing sessions of this transactions
func SpendingFingerprint(g *ReversibleGraph, uid string) ([]FingerPrint, int, error) {
	// maximumDistance is the maximum distance between to earliest (lowest) input timestamp
	// of the root transaction and the timestamp of the compared transaction
	// 2 days = 60 * 60 * 24 * 2 = 172800 seconds
	const maximumDistance = 172800

	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, 0, err
	}

	rootNode := g.Node(nodeUID)
	if rootNode == nil {
		return nil, 0, serror.NewStackErrorStr(uid + " not in graph")
	}

	rootTx, ok := rootNode.(TransactionNode)
	if !ok || !rootTx.PrivacyType.IsDestination() {
		return nil, 0, serror.NewStackErrorStr(uid + " is not a destination transaction")
	}

	rootMeans := getSessionsFromTransaction(g, rootTx)
	if len(rootMeans) == 0 {
		return nil, 0, serror.NewStackErrorStr(uid + " has no session means")
	}
	earliestInputTimestamp := rootMeans[0]
	numSessions := len(rootMeans)
	numSessionsFloat := float64(numSessions)
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
		if !ok || !txNode.PrivacyType.IsDestination() || earliestInputTimestamp-txNode.TS.Unix() > maximumDistance {
			continue
		}

		txMeans := getSessionsFromTransaction(g, txNode)
		if len(txMeans) == 0 {
			continue
		}

		// calculate score and check if score is high enough
		score, err := scoreMeans(rootMeans, txMeans)
		if err != nil || score < 0.7 || score/numSessionsFloat < 0.5 {
			continue
		}

		fingerprints = append(fingerprints, FingerPrint{
			TransactionUID: txNode.String(),
			Score:          score / numSessionsFloat,
			SessionCount:   numSessions,
		})
		sort.Slice(fingerprints, func(i, j int) bool {
			return fingerprints[i].Score < fingerprints[j].Score
		})

		// remove the first element (which has the lowest score)
		if len(fingerprints) > maxNumberOfScoreResults {
			fingerprints = fingerprints[1:]
		}
	}

	return fingerprints, numSessions, err
}
