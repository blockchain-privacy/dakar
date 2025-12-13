// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"backend/analytics/heuristics"
	"backend/db"
	"backend/server"
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

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

func (s *Server) listHeuristics() mcp.ToolHandlerFor[any, *ListHeuristicsResult] {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *ListHeuristicsResult, error) {
		result := ListHeuristicsResult{Descriptors: make([]heuristics.Descriptor, 0, len(heuristics.ConstructorMap))}
		for _, v := range heuristics.ConstructorMap {
			result.Descriptors = append(result.Descriptors, v().GetDescriptor())
		}
		return nil, &result, nil
	}
}

func (s *Server) executeHeuristic() mcp.ToolHandlerFor[heuristics.HeuristicOptions, *ExecuteHeuristicResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, opt heuristics.HeuristicOptions) (*mcp.CallToolResult, *ExecuteHeuristicResult, error) {
		tUser, err := server.ExtractTokenUser(ctx)
		if err != nil {
			return nil, nil, err
		}

		info("received options", "options", opt)

		parentUID, err := db.GetTransactionUID(ctx, s.db, opt.TransactionHash)
		if err != nil {
			return nil, nil, err
		}

		if !opt.IsValid(ctx, s.db, parentUID) {
			return nil, nil, serror.FromStr("invalid options")
		}

		executor, err := heuristics.ConstructExecutors(opt, tUser.ID, parentUID)
		if err != nil {
			return nil, nil, err
		}

		w := heuristicWork{executor: executor}

		done := s.worker.AddWork(ctx, &w)
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

		info("results", "clusters", w.clusters[:min(len(w.clusters), 10)])

		return nil, &ExecuteHeuristicResult{ResultCount: len(w.clusters)}, nil
	}
}
