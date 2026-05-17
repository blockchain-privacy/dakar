// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	dbh "backend/db/heuristics"
	"backend/external"
	"backend/workspace"
	"context"
)

type TransactionParams struct {
	TransactionHash string `json:"transactionHash" jsonschema:"required"`
	GetProperty     string `json:"getProperty" jsonschema:"required, allowed values: all (gets all details), type (gets only the transaction type)"`
}

type ListHeuristicsResult struct {
	Descriptors []heuristics.Descriptor `json:"descriptors,omitempty" jsonschema:"the descriptors of all possible heuristics"`
}

type ExecuteHeuristicParam struct {
	Options []heuristics.HeuristicOptions `json:"options,omitempty" jsonschema:"heuristics that will be run in sequential order. Each heuristic receives the results of the previous one."`
}

type ExecuteHeuristicResult struct {
	Counts []int `json:"counts,omitempty" jsonschema:"each item contains the number of clusters found by the heuristic. Ordered by heuristic run order"`
}

type heuristicWork struct {
	// h is the heuristic which will get executed
	h             heuristics.Heuristic
	parentUID     string
	parentResults []dbh.HeuristicCluster
	results       []dbh.HeuristicCluster
}

func (d *heuristicWork) Run(ctx context.Context, _ *workspace.Mutex, c external.Database, g *graph.Wrapper) error {
	var err error
	d.results, err = d.h.Exec(ctx, c, g, d.parentUID, d.parentResults)
	if err != nil {
		return err
	}

	return nil
}
