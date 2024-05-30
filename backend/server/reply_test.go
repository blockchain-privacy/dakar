package server

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func Test_getTransactionReply(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	tests := []struct {
		query      string
		wantReply  transactionReply
		wantStatus int
	}{
		{
			query:      "1",
			wantStatus: http.StatusNotFound,
		},
		{
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			query:      "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
			wantStatus: 0,
			wantReply: transactionReply{
				Transactions: []db.FrontendTransaction{
					{Hash: "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb"},
				},
			},
		},
	}
	for _, tt := range tests {
		reply, status := getTransactionReply(dbHandle, tt.query)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Len(t, reply.Transactions, len(tt.wantReply.Transactions))
			require.Equal(t, tt.wantReply.Transactions[0].Hash, reply.Transactions[0].Hash)
		}
	}
}

func Test_getBlockReply(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	tests := []struct {
		query      string
		request    *http.Request
		wantReply  blockReply
		wantStatus int
	}{
		{
			query:      "1",
			request:    httptest.NewRequest(http.MethodGet, "/blockchain/blocks?offset=asdf", nil),
			wantStatus: http.StatusBadRequest,
		},
		{
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			query:      "1",
			request:    httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil),
			wantStatus: http.StatusNotFound,
		},
		{
			query:   "60000",
			request: httptest.NewRequest(http.MethodGet, "/blockchain/blocks", nil),
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
		{
			query:   "60000",
			request: httptest.NewRequest(http.MethodGet, "/block?offset=5", nil),
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
		{
			query:   "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
			request: httptest.NewRequest(http.MethodGet, "/block?offset=5", nil),
			wantReply: blockReply{Block: &db.FrontendBlock{
				Hash:             "000000000013629708e60a0a20c2161a1195f8ba4871eaf408baf847bca84f71",
				ID:               60000,
				TransactionCount: 2,
			}},
			wantStatus: 0,
		},
	}
	for _, tt := range tests {
		reply, status := getBlockReply(tt.request, dbHandle, tt.query)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Equal(t, tt.wantReply.Block.Hash, reply.Block.Hash)
			require.Equal(t, tt.wantReply.Block.ID, reply.Block.ID)
			require.Equal(t, tt.wantReply.Block.TransactionCount, reply.Block.TransactionCount)
		}
	}
}

func Test_getShortestTransactionPathReply(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	tests := []struct {
		body       io.Reader
		wantReply  shortestTransactionPathReply
		wantStatus int
	}{
		// invalid json
		{
			body: strings.NewReader(`{
				"from": "asdf",
				"to": 
			}`),
			wantStatus: http.StatusBadRequest,
		},
		// equal values
		{
			body: strings.NewReader(`{
				"from": "asdf",
				"to": "asdf"
			}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			body: strings.NewReader(`{
				"from": "asdf1",
				"to": "asdf2"
			}`),
			wantStatus: http.StatusNotFound,
		},
		{
			body: strings.NewReader(`{
				"from": "818dae776566815b8d5307f8597fc8c1db737e933a4605e1841a83f078731638",
				"to": "18aa3626fe0f46d15d14a8044bda0f479d8b5cff8295fd24fbebccd449cb7eb4"
			}`),
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
			body: strings.NewReader(`{
				"from": "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb",
				"to": "ae52511e1f61977ee2993e47f387d6fe409140dee5783f6df07703360c81a542",
				"includePrivacyTransactions": true,
				"anyDirection": true
			}`),
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
		reply, status := getShortestTransactionPathReply(dbHandle, tt.body)
		require.Equal(t, tt.wantStatus, status)
		if status == http.StatusOK || status == 0 {
			require.Equal(t, len(tt.wantReply.Transactions), len(reply.Transactions))
		}
	}
}
