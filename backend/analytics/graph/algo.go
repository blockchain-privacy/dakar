package graph

import (
	"github.com/qrest/gomisc/serror"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
	"strconv"
	"time"
)

func ErrNodeNotFound(nodeID int64) error {
	return serror.FromFormat("node %s does not exist in graph", ToHex(nodeID))
}

// ToHex returns a hexadecimal string representation of the given integer with the '0x' prefix.
// Does not handle negative integers.
func ToHex(i int64) string {
	return "0x" + strconv.FormatInt(i, 16)
}

// ToInteger a hex string in the form of "0x123" to an integer. For performance reasons, there is no strong input validation.
func ToInteger(hexString string) (int64, error) {
	if len(hexString) < 3 {
		return 0, serror.FromFormat("string to short: %s", hexString)
	}
	integer, err := strconv.ParseInt(hexString[2:], 16, 64)
	if err != nil {
		return 0, serror.New(err)
	}
	return integer, nil
}

// HasSpendingGap returns true if the spending gap is bigger than 8 hours
func HasSpendingGap(g *ReversibleGraph, edge AddressEdge) bool {
	fromTS := g.Node(edge.F.ID()).(TransactionNode).TS
	toTS := g.Node(edge.T.ID()).(TransactionNode).TS

	if g.reversed {
		return toTS.Sub(fromTS) >= time.Hour*8
	}

	return fromTS.Sub(toTS) >= time.Hour*8
}

// CheckAddressExclusions returns true if any of the addresses of the given edge
// are not contained in the exclusion map, or if the given map is empty
func CheckAddressExclusions(exclusions map[int64]bool, edge AddressEdge) bool {
	if len(exclusions) == 0 {
		return true
	}

	for _, address := range edge.AddressUIDs {
		if !exclusions[address] {
			return true
		}
	}

	return false
}

// ReverseLookupByID performs a reverse lookup starting at the given nodeId
func ReverseLookupByID(g *ReversibleGraph, nodeID int64, maxLookBackTime time.Duration,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	node := g.Node(nodeID)
	if node == nil {
		return nil, ErrNodeNotFound(nodeID)
	}

	foundEndpoints := make(map[string]bool)

	nodeTS := node.(TransactionNode).TS

	isReversed := g.IsReversed()

	exclusionsMap := make(map[int64]bool, len(addressExclusions))

	for _, e := range addressExclusions {
		integer, err := ToInteger(e)
		if err != nil {
			return nil, err
		}

		exclusionsMap[integer] = true
	}

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			addressEdge := e.(AddressEdge)
			if excludeSpendingGaps && HasSpendingGap(g, addressEdge) {
				return false
			}

			if !CheckAddressExclusions(exclusionsMap, addressEdge) {
				return false
			}

			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(TransactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				// isReversed is true if it is a forward lookup: default case is a reverse
				// lookup so if the graph is reversed a forward lookup is happening
				if isReversed {
					if toNode.TS.Sub(nodeTS) > maxLookBackTime {
						// The forward lookup starts at an origin transaction and looks forward
						// for a certain user-defined duration. Funds from mixing transaction are usually
						// not spent directly after the mixing process is finished. Because of this, the next
						// connected non-mixing transaction must be included to the returned set.
						if !toNode.PrivacyType.IsMixing() {
							foundEndpoints[toNode.String()] = true
						}

						return false
					}
				} else if nodeTS.Sub(toNode.TS) > maxLookBackTime {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if !toNode.PrivacyType.IsMixing() {
				foundEndpoints[toNode.String()] = true
				return false
			}

			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	w.Walk(g, node, func(n graph.Node, _ int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			thisNode := n.(TransactionNode)
			foundEndpoints[thisNode.String()] = true
		}

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})

	return foundEndpoints, nil
}

// ReverseLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph backward
func ReverseLookup(g *ReversibleGraph, uid string, maxLookBackTime time.Duration,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)

	return ReverseLookupByID(g, nodeUID, maxLookBackTime, addressExclusions, excludeSpendingGaps)
}

// ForwardLookup returns all leaf nodes of the tree which has uid as its root while traversing the graph forward.
// It does not traverse paths which have a timestamp younger than the node specified by targetUid.
func ForwardLookup(g *ReversibleGraph, uid string, targetUID string,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}

	targetNodeUID, err := ToInteger(targetUID)
	if err != nil {
		return nil, err
	}

	node := g.Node(nodeUID).(TransactionNode)
	targetNode := g.Node(targetNodeUID).(TransactionNode)

	g.SetReverse(true)
	origins, err := ReverseLookupByID(g, nodeUID, targetNode.TS.Sub(node.TS), addressExclusions, excludeSpendingGaps)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

// ForwardLookupByTime returns all leaf nodes of the tree which has uid as its root while traversing the graph forward
// It does not traverse paths which are outside maxLookForwardTime.
func ForwardLookupByTime(g *ReversibleGraph, uid string, maxLookForwardTime time.Duration,
	addressExclusions []string, excludeSpendingGaps bool) (map[string]bool, error) {
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}

	g.SetReverse(true)
	origins, err := ReverseLookupByID(g, nodeUID, maxLookForwardTime, addressExclusions, excludeSpendingGaps)
	if err != nil {
		return nil, err
	}
	g.SetReverse(false)
	return origins, err
}

// GetInputTransactions returns the uids of all directly connected input transactions of the tx specified by uid
func GetInputTransactions(g *ReversibleGraph, uid string) ([]string, error) {
	// convert hex string to integer
	nodeUID, err := ToInteger(uid)
	if err != nil {
		return nil, err
	}

	// check if node exists
	if g.Node(nodeUID) == nil {
		return nil, ErrNodeNotFound(nodeUID)
	}

	var uids []string
	fromNodes := g.From(nodeUID)
	for fromNodes.Next() {
		uids = append(uids, fromNodes.Node().(TransactionNode).String())
	}

	return uids, nil
}
