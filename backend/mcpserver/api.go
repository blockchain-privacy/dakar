// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
	"gitlab.com/blockchain-privacy/dakar/db"
	dbh "gitlab.com/blockchain-privacy/dakar/db/heuristics"
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

func (s *Server) executeHeuristic() mcp.ToolHandlerFor[ExecuteHeuristicParam, *ExecuteHeuristicResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, opt ExecuteHeuristicParam) (*mcp.CallToolResult, *ExecuteHeuristicResult, error) {
		info("received options", "options", opt)

		if len(opt.Options) == 0 {
			return nil, nil, serror.FromStr("received no heuristic options")
		}

		if opt.Options[0].TransactionHash == "" {
			return nil, nil, serror.FromStr("heuristic option does not contain transaction hash")
		}

		var parentResults []dbh.HeuristicCluster
		var result ExecuteHeuristicResult
		for i, options := range opt.Options {
			w := heuristicWork{}
			// first heuristic needs to validate its parent (transaction)
			if i == 0 {
				parentUID, err := db.GetTransactionUID(ctx, s.db, options.TransactionHash)
				if err != nil {
					return nil, nil, err
				}

				if !options.IsValid(ctx, s.db, parentUID) {
					return nil, nil, serror.FromStr("invalid options")
				}
				w.parentUID = parentUID
			} else {
				if !options.CheckParameterAndType() {
					return nil, nil, serror.FromStr("invalid options")
				}
				w.parentResults = parentResults
			}

			h, err := options.CreateHeuristic()
			if err != nil {
				return nil, nil, err
			}

			w.h = h

			done := s.worker.AddWork(ctx, &w)
			if done == nil {
				return nil, nil, serror.FromStr("could not add work")
			}

			select {
			case <-ctx.Done():
				return nil, nil, serror.FromStr("context timed out")
			case err := <-done:
				if err != nil {
					warn(err)
					return nil, nil, err
				}
			}

			info("results", "options", options, "clusters", w.results[:min(len(w.results), 10)])

			result.Counts = append(result.Counts, len(w.results))

			if len(w.results) == 0 {
				// stop, because there is nothing to do for the next heuristic
				break
			}

			// save results so they can be passed to the next heuristic
			parentResults = w.results
		}

		return nil, &result, nil
	}
}
