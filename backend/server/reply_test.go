// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"backend/db"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getTransactionReply(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	tests := []struct {
		setRequest func()
		wantReply  transactionReply
		wantStatus int
	}{
		{
			setRequest: func() { r.SetPathValue("hash", "1") },
			wantStatus: http.StatusNotFound,
		},
		{
			setRequest: func() { r.SetPathValue("hash", "") },
			wantStatus: http.StatusBadRequest,
		},
		{
			setRequest: func() { r.SetPathValue("hash", "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb") },
			wantStatus: 0,
			wantReply: transactionReply{
				Transactions: []db.FrontendTransaction{
					{Hash: "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb"},
				},
			},
		},
	}
	for _, tt := range tests {
		tt.setRequest()
		reply, status := getTransactionReply(dbHandle, r)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Len(t, reply.Transactions, len(tt.wantReply.Transactions))
			require.Equal(t, tt.wantReply.Transactions[0].Hash, reply.Transactions[0].Hash)
		}
	}
}

func Test_getBlockReply(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	tests := []struct {
		getRequest func() *http.Request
		wantReply  blockReply
		wantStatus int
	}{
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/blockchain/blocks?offset=asdf", nil)
				r.SetPathValue("hash", "1")
				return r
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)
				r.SetPathValue("hash", "")
				return r
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)
				r.SetPathValue("hash", "1")
				return r
			},
			wantStatus: http.StatusNotFound,
		},
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil)
				r.SetPathValue("hash", "60000")
				return r
			},
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/block?offset=5", nil)
				r.SetPathValue("hash", "60000")
				return r
			},
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
		{
			getRequest: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/block?offset=5", nil)
				r.SetPathValue("hash", "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71")
				return r
			},
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
	}
	for _, tt := range tests {
		r := tt.getRequest()
		reply, status := getBlockReply(dbHandle, r)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Equal(t, tt.wantReply.Block.Hash, reply.Block.Hash)
			require.Equal(t, tt.wantReply.Block.ID, reply.Block.ID)
			require.Equal(t, tt.wantReply.Block.TransactionCount, reply.Block.TransactionCount)
		}
	}
}

func Test_getShortestTransactionPathReply(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	tests := []struct {
		r          *http.Request
		wantReply  shortestTransactionPathReply
		wantStatus int
	}{
		// invalid json
		{
			r: httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{
				"from": "asdf",
				"to": 
			}`)),
			wantStatus: http.StatusBadRequest,
		},
		// equal values
		{
			r: httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{
				"from": "asdf",
				"to": "asdf"
			}`)),
			wantStatus: http.StatusBadRequest,
		},
		{
			r: httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{
				"from": "asdf1",
				"to": "asdf2"
			}`)),
			wantStatus: http.StatusNotFound,
		},
		{
			r: httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{
				"from": "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638",
				"to": "18aa3626fe0f46d15d14a8044bda0f479d8b5cff8295fd24fbebccd449cb7eb4"
			}`)),
			wantReply: shortestTransactionPathReply{
				Transactions: []db.FrontendTransaction{
					{Hash: "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638"},
					{Hash: "af25e5385300cfbec9ecba1e7c75035b1c1e77853250db08ac7e455476f5c310"},
					{Hash: "18aa3626fe0f46d15d14a8044bda0f479d8b5cff8295fd24fbebccd449cb7eb4"},
				},
			},
			wantStatus: 0,
		},
		{
			r: httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{
				"from": "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
				"to": "ae52511e1f61977ee2993e47f387d6fe409140dee5783f6df07703360c81a542",
				"includePrivacyTransactions": true,
				"anyDirection": true
			}`)),
			wantReply: shortestTransactionPathReply{
				Transactions: []db.FrontendTransaction{
					{Hash: "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb"},
					{Hash: "ae52511e1f61977ee2993e47f387d6fe409140dee5783f6df07703360c81a542"},
				},
			},
			wantStatus: 0,
		},
	}
	for _, tt := range tests {
		reply, status := getShortestTransactionPathReply(dbHandle, tt.r)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Len(t, reply.Transactions, len(tt.wantReply.Transactions))
		}
	}
}
