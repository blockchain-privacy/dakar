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

type ExecuteHeuristicResult struct {
	ResultCount int `json:"resultCount,omitempty" jsonschema:"the number of transactions found by the heuristic"`
}

type heuristicWork struct {
	executor heuristics.Executor
	clusters []dbh.HeuristicCluster
}

func (d *heuristicWork) Run(ctx context.Context, _ *workspace.Mutex, c external.Database, g *graph.Wrapper) error {
	var err error
	d.clusters, err = d.executor.Run(ctx, c, g)
	if err != nil {
		return err
	}

	return nil
}
