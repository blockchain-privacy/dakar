// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	mgraph "backend/analytics/graph"
	"backend/constants"
	"backend/db/analytics/exclusion"
	"backend/external"
	"context"
	"fmt"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"
	"math/rand/v2"
	"time"
)

func doSimulation(ctx context.Context, database external.Database, g *mgraph.ReversibleGraph,
	nodeIDStr string, userUID string, lookBackTimeHours int) {
	exclusionMap, err := getExclusionMap(ctx, database, userUID)
	if err != nil {
		warn(err)
		return
	}

	nodeID, err := mgraph.ToInteger(nodeIDStr)
	if err != nil {
		warn(err)
		return
	}

	maxLookBackTime := time.Hour * time.Duration(lookBackTimeHours)

	info(fmt.Sprintf("Random simulation for %s starting", maxLookBackTime.String()))
	err = doRandomSimulation(g, nodeID, maxLookBackTime, exclusionMap)
	if err != nil {
		warn(err)
		return
	}

	info(fmt.Sprintf("Random simulation for %s ended", maxLookBackTime.String()))

	info(fmt.Sprintf("Semi simulation for %s starting", maxLookBackTime.String()))
	err = doSemiRandomSimulation(g, nodeID, maxLookBackTime, exclusionMap)
	if err != nil {
		warn(err)
		return
	}
	info(fmt.Sprintf("Semi simulation for %s ended", maxLookBackTime.String()))

	info(fmt.Sprintf("Semi simulation V2 for %s starting", maxLookBackTime.String()))
	err = doSemiRandomSimulationV2(g, nodeID, maxLookBackTime, exclusionMap)
	if err != nil {
		warn(err)
		return
	}
	info(fmt.Sprintf("Semi simulation V2 for %s ended", maxLookBackTime.String()))
}

func getExclusionMap(ctx context.Context, c external.Database, userID string) (exclusions map[int64]bool, err error) {
	exclusionSlice, err := exclusion.GetAddressExclusionUIDs(ctx, c, userID)
	if err != nil {
		return nil, err
	}

	exclusionsMap := make(map[int64]bool, len(exclusionSlice))

	for _, e := range exclusionSlice {
		integer, err := mgraph.ToInteger(e)
		if err != nil {
			return nil, err
		}

		exclusionsMap[integer] = true
	}

	return exclusionsMap, nil
}

// getNodeNumbers performs a reverse lookup starting at the
// given nodeId and collects statistics about traversed mixing nodes, result nodes and addresses
func getNodeNumbers(g *mgraph.ReversibleGraph, nodeID int64, maxLookBackTime time.Duration,
	exclusions map[int64]bool) (addressesPerTransactions [][]int64, addressMap map[int64]bool, endpoints map[int64]bool,
	mixingTxs map[int64]bool, err error) {
	node := g.Node(nodeID)
	if node == nil {
		err = mgraph.ErrNodeNotFound(nodeID)
		return
	}

	nodeTS := node.(mgraph.TransactionNode).TS

	isReversed := g.IsReversed()

	globalEndpointMap := make(map[int64]bool)

	w := traverse.BreadthFirst{
		Traverse: func(e graph.Edge) bool {
			if !mgraph.CheckAddressExclusions(exclusions, e.(mgraph.AddressEdge)) {
				return false
			}

			// get node to which the edge leads
			toNode := g.Node(e.To().ID()).(mgraph.TransactionNode)

			// if a maximum look back time is set check the timestamp
			if maxLookBackTime > 0 {
				// isReversed is true if it is a forward lookup: default case is a reverse
				// lookup so if the graph is reversed a forward lookup is happening
				if isReversed {
					if toNode.TS.Sub(nodeTS) > maxLookBackTime {
						return false
					}
				} else if nodeTS.Sub(toNode.TS) > maxLookBackTime {
					return false
				}
			}

			// if it is not a mixing transaction save it and stop following that edge
			if toNode.Type != constants.TypeDashMixing {
				globalEndpointMap[toNode.ID()] = true
				return false
			}

			// true: follow this link
			// false: do not follow this link
			return true
		},
	}

	globalAddressMap := make(map[int64]bool)
	globalMixingMap := make(map[int64]bool)

	w.Walk(g, node, func(n graph.Node, _ int) bool {
		from := g.From(n.ID())
		if from.Len() == 0 {
			globalEndpointMap[n.ID()] = true
		}

		if g.Node(n.ID()).(mgraph.TransactionNode).Type == constants.TypeDashMixing {
			globalMixingMap[n.ID()] = true
		}

		var addresses []int64
		for from.Next() {
			edgeAddresses := g.Edge(n.ID(), from.Node().ID()).(mgraph.AddressEdge).AddressUIDs
			for _, a := range edgeAddresses {
				globalAddressMap[a] = true
				addresses = append(addresses, a)
			}
		}

		addressesPerTransactions = append(addressesPerTransactions, addresses)

		// true: stop traversing nodes
		// false: do not stop traversing nodes
		return false
	})

	addressMap = globalAddressMap
	endpoints = globalEndpointMap
	mixingTxs = globalMixingMap

	return
}

// getSimulatedRandomExclusionSet returns a specified percentage of address IDs as a map
func getSimulatedRandomExclusionSet(addressMap map[int64]bool, percentExcluded float64) map[int64]bool {
	if percentExcluded <= 0.0 || percentExcluded >= 1.0 {
		return nil
	}

	exclusionsMap := make(map[int64]bool)

	addresses := make([]int64, 0, len(addressMap))

	for a := range addressMap {
		addresses = append(addresses, a)
	}
	numAddresses := len(addresses)

	for float64(len(exclusionsMap))/float64(numAddresses) < percentExcluded {
		exclusionsMap[addresses[rand.IntN(numAddresses)]] = true //nolint:gosec
	}

	return exclusionsMap
}

// getSimulatedSemiRandomExclusionSet returns a specified percentage of address IDs as a map.
// Depending on the number of participants, a number of inputs is chosen from each transaction.
// It is assumed that there are 3 participants per transaction.
// Therefore, if participant == 2 and a transaction has 9 inputs, either 0, 3 or 6 inputs will be selected.
func getSimulatedSemiRandomExclusionSet(transactionAddresses [][]int64, participants int,
	percentExcluded float64) map[int64]bool {
	if percentExcluded <= 0.0 || percentExcluded >= 1.0 || participants <= 0 {
		return nil
	}

	exclusionsMap := make(map[int64]bool)

	allAddresses := make(map[int64]bool)

	for _, t := range transactionAddresses {
		for _, input := range t {
			allAddresses[input] = true
		}
	}

	numAddresses := len(allAddresses)
	numTransactions := len(transactionAddresses)
	usedTransactionIndex := make(map[int]bool)

	maxBoundParticipants := 3

	if participants < 3 {
		maxBoundParticipants = participants
	}

	for float64(len(exclusionsMap))/float64(numAddresses) < percentExcluded {
		if len(usedTransactionIndex) == len(transactionAddresses) {
			break
		}

		index := rand.IntN(numTransactions) //nolint:gosec
		if usedTransactionIndex[index] {
			continue
		}

		addresses := transactionAddresses[index]
		usedTransactionIndex[index] = true

		// get the number of participants we simulate the data from.
		// This should not be always the same number for participants > 1, so we randomize.
		chosenParticipants := 1
		if participants > 1 {
			randInt := rand.IntN(100) //nolint:gosec

			if randInt > 80 && randInt <= 95 {
				chosenParticipants = 2
			} else if maxBoundParticipants > 2 && randInt > 95 {
				chosenParticipants = 3
			}
		}

		// the number of inputs of this transaction we collect for the exclusion list
		numInputs := len(addresses) / chosenParticipants

		if numInputs == 0 {
			continue
		}

		collectedAddresses := make(map[int64]bool)

		for len(collectedAddresses) < numInputs {
			collectedAddresses[addresses[rand.IntN(len(addresses))]] = true //nolint:gosec
		}

		for c := range collectedAddresses {
			exclusionsMap[c] = true
		}
	}

	return exclusionsMap
}

// getSimulatedSemiRandomExclusionSetV2 returns the IDs of addresses fitting to the simulated parameters.
// It is assumed that there are 3 participants per transaction.
func getSimulatedSemiRandomExclusionSetV2(transactionAddresses [][]int64, participants int,
	mixingParticipants int) map[int64]bool {
	if mixingParticipants <= 0 || participants <= 0 {
		return nil
	}

	exclusionsMap := make(map[int64]bool)

	for _, tx := range transactionAddresses {
		// randomly select participants
		selectedParticipants := make(map[int]bool)
		for len(selectedParticipants) < 3 {
			selectedParticipants[rand.IntN(mixingParticipants)+1] = true //nolint:gosec
		}

		// calculate the number of participants we have data from
		var numParticipantsWithData float64
		for p := range selectedParticipants {
			if p <= participants {
				numParticipantsWithData++
			}
		}

		if numParticipantsWithData == 0 {
			// no data for this transaction
			continue
		}

		if numParticipantsWithData == 3 {
			// all addresses are selected
			for _, address := range tx {
				exclusionsMap[address] = true
			}
			continue
		}

		numSelectedAddresses := (int)(float64(len(tx)) / 3.0 * numParticipantsWithData)

		selectedAddresses := make(map[int64]bool)
		for len(selectedAddresses) < numSelectedAddresses {
			selectedAddresses[tx[rand.IntN(len(tx))]] = true //nolint:gosec
		}

		for s := range selectedAddresses {
			exclusionsMap[s] = true
		}
	}

	return exclusionsMap
}

func doRandomSimulation(g *mgraph.ReversibleGraph, nodeID int64, maxLookBackTime time.Duration,
	addressExclusions map[int64]bool) error {
	_, usedAddresses, endpoints, mixingTxs, err := getNodeNumbers(g, nodeID, maxLookBackTime, nil)
	if err != nil {
		return err
	}

	if len(usedAddresses) > 0 {
		info("Node numbers without real filter", "addresses", len(usedAddresses),
			"endpoints", len(endpoints), "mixing txs", len(mixingTxs),
			"endpoint reduction", 1-float64(len(endpoints))/float64(len(endpoints)))

		simulations := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}

		for _, s := range simulations {
			simulatedExclusions := getSimulatedRandomExclusionSet(usedAddresses, s)
			_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, simulatedExclusions)
			if err != nil {
				return err
			}

			info(fmt.Sprintf("Simulated %f node numbers with filter", s), "addresses", len(a), "endpoints",
				len(e), "mixing txs", len(m), "endpoint reduction", 1-float64(len(e))/float64(len(endpoints)))
		}

		_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, addressExclusions)
		if err != nil {
			return err
		}
		info("Node numbers with real filter", "addresses", len(a), "endpoints", len(e), "mixing txs", len(m),
			"endpoint reduction", 1-float64(len(e))/float64(len(endpoints)))
	}

	return nil
}

func doSemiRandomSimulation(g *mgraph.ReversibleGraph, nodeID int64, maxLookBackTime time.Duration,
	addressExclusions map[int64]bool) error {
	transactionAddresses, usedAddresses, endpoints, mixingTxs, err := getNodeNumbers(g, nodeID, maxLookBackTime, nil)
	if err != nil {
		return err
	}

	if len(usedAddresses) > 0 {
		info("Node numbers without real filter", "addresses", len(usedAddresses),
			"endpoints", len(endpoints), "mixing txs", len(mixingTxs),
			"endpoint reduction", 1-float64(len(endpoints))/float64(len(endpoints)))

		simulationParticipants := []int{1, 2, 3}
		simulationPercentages := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}

		for _, p := range simulationParticipants {
			for _, s := range simulationPercentages {
				simulatedExclusions := getSimulatedSemiRandomExclusionSet(transactionAddresses, p, s)
				_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, simulatedExclusions)
				if err != nil {
					return err
				}
				info("Simulated", "participants", p, "percentage", s,
					"Node numbers with filter: addresses", len(a), "endpoints",
					len(e), "mixing txs", len(m), "endpoint reduction",
					1-float64(len(e))/float64(len(endpoints)))
			}
		}

		_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, addressExclusions)
		if err != nil {
			return err
		}
		info("Node numbers with real filter", "addresses", len(a), "endpoints", len(e), "mixing txs", len(m),
			"endpoint reduction", 1-float64(len(e))/float64(len(endpoints)))
	}

	return nil
}

func doSemiRandomSimulationV2(g *mgraph.ReversibleGraph, nodeID int64, maxLookBackTime time.Duration,
	addressExclusions map[int64]bool) error {
	transactionAddresses, usedAddresses, endpoints, mixingTxs, err := getNodeNumbers(g, nodeID, maxLookBackTime, nil)
	if err != nil {
		return err
	}

	if len(usedAddresses) > 0 {
		info("Node numbers without real filter", "addresses", len(usedAddresses), "endpoints",
			len(endpoints), "mixing txs", len(mixingTxs),
			"endpoint reduction", 1-float64(len(endpoints))/float64(len(endpoints)))

		simulationParticipants := []int{1, 2, 3, 4, 5}
		simulationMixingParticipants := []int{5, 10, 15, 20, 25, 30}

		for _, p := range simulationParticipants {
			for _, s := range simulationMixingParticipants {
				simulatedExclusions := getSimulatedSemiRandomExclusionSetV2(transactionAddresses, p, s)
				_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, simulatedExclusions)
				if err != nil {
					return err
				}
				info("Simulated", "participants", p, "global mixing participants", s,
					"Node numbers with filter: addresses", len(a), "endpoints",
					len(e), "mixing txs", len(m), "endpoint reduction",
					1-float64(len(e))/float64(len(endpoints)))
			}
		}

		_, a, e, m, err := getNodeNumbers(g, nodeID, maxLookBackTime, addressExclusions)
		if err != nil {
			return err
		}
		info("Node numbers with real filter", "addresses", len(a), "endpoints", len(e), "mixing txs", len(m),
			"endpoint reduction", 1-float64(len(e))/float64(len(endpoints)))
	}

	return nil
}
