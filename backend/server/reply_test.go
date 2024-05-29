package server

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
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
