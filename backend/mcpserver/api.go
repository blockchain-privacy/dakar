// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"backend/db"
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
