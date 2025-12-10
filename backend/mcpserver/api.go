// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"backend/analytics/graph"
	"backend/analytics/heuristics"
	"backend/db"
	"backend/external"
	"backend/workspace"
	"context"
	"errors"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type TransactionParams struct {
	TransactionHash string `json:"transactionHash" jsonschema:"required"`
	GetProperty     string `json:"getProperty" jsonschema:"required, allowed values: all (gets all details), type (gets only the transaction type)"`
}

func (s *Server) getTransaction() mcp.ToolHandlerFor[TransactionParams, *db.FrontendTransaction] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input TransactionParams) (*mcp.CallToolResult, *db.FrontendTransaction, error) {
		transactions, err := db.GetFrontendTransaction(ctx, s.db, input.TransactionHash)
		if err != nil {
			// only print error if it is not expected
			if !errors.Is(err, db.ErrTransactionNotFound) {
				warn(err)
			}

			return nil, nil, err
		}

		info("received", "input", input)

		if input.GetProperty == "type" {
			var result db.FrontendTransaction
			result.Type = transactions[0].Type
			result.Hash = transactions[0].Hash
			return nil, &result, nil
		}

		return nil, &transactions[0], nil
	}
}

type ListHeuristicsResult struct {
	Descriptors []heuristics.Descriptor `json:"descriptors,omitempty" jsonschema:"the descriptors of all possible heuristics"`
}

func (s *Server) listHeuristics() mcp.ToolHandlerFor[any, *ListHeuristicsResult] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *ListHeuristicsResult, error) {
		result := ListHeuristicsResult{Descriptors: make([]heuristics.Descriptor, 0, len(heuristics.ConstructorMap))}
		for _, v := range heuristics.ConstructorMap {
			result.Descriptors = append(result.Descriptors, v().GetDescriptor())
		}
		return nil, &result, nil
	}
}

type ExecuteHeuristicParams struct {
	Type            string `json:"type,omitempty" jsonschema:"required, the type of the heuristic"`
	Parameter       string `json:"parameter,omitempty" jsonschema:"required, the heuristic parameter"`
	TransactionHash string `json:"transactionHash,omitempty" jsonschema:"required, the transaction for that the transaction is created for"`
}

type ExecuteHeuristicResult struct {
	ResultCount int `json:"resultCount,omitempty" jsonschema:"the number of transactions found by the heuristic"`
}

type dummyWork struct {
}

func (d *dummyWork) Run(ctx context.Context, _ *workspace.Mutex, _ external.Database, _ *graph.Wrapper) error {
	ticker := time.Tick(time.Second * 60 * 5)

	select {
	case <-ctx.Done():
		info("request context done")
		return serror.FromStr("context done")
	case <-ticker:
		info("request ticker done")
	}

	return nil
}

func (s *Server) executeHeuristic() mcp.ToolHandlerFor[ExecuteHeuristicParams, *ExecuteHeuristicResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ ExecuteHeuristicParams) (*mcp.CallToolResult, *ExecuteHeuristicResult, error) {
		info("test")
		done := s.worker.AddWork(ctx, &dummyWork{})
		if done == nil {
			return nil, nil, serror.FromStr("could not add work")
		}

		select {
		case <-ctx.Done():
			info("context timed out")
			return nil, nil, serror.FromStr("context timed out")
		case err := <-done:
			info("finished work")
			if err != nil {
				warn(err)
				return nil, nil, err
			}
		}

		return nil, &ExecuteHeuristicResult{ResultCount: 10}, nil
	}
}
