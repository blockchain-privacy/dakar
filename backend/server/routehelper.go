// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"gitlab.com/blockchain-privacy/dakar/analytics/heuristics"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	dbstat "gitlab.com/blockchain-privacy/dakar/db/status"
	"gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// isValidInput is a regex filter which checks if the input only consists of numbers and letters
var isValidInput = regexp.MustCompile(`^[a-zA-Z\d]*$`).MatchString

// isValid checks if user input is valid.
// Should be used to check address, transaction and block hashes, as well as block ids.
func isValid(input string) bool {
	inputLen := len(input)
	// 64 -> length of transaction hash and block hash
	if inputLen == 0 || inputLen > 64 {
		return false
	}

	// 26 -- 62 -> address length; if smaller than it must be a block id
	if inputLen < 26 {
		// attempt to convert input to an integer; if it succeeds the input is valid.
		_, err := strconv.Atoi(input)
		return err == nil
	}

	return isValidInput(input)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
}

// SendReply encodes the given reply into JSON and sends it
func SendReply(w http.ResponseWriter, reply any, statusCode int) {
	setCORSHeaders(w)

	// use marshalling instead of encoding (streaming), as it gives better error handling
	// and because encoding buffers all data before writing: https://github.com/golang/go/issues/7872
	// todo check if https://github.com/golang/go/discussions/63397 has been released and then rework json handling.
	replyBuffer, err := json.Marshal(reply)
	if err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
		warn(serror.New(err))
		return
	}

	if reply == "" {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}

	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	if _, err := w.Write(replyBuffer); err != nil {
		// not possible to send response to client, so just log error
		warn(serror.New(err))
	}
}

type searchReply struct {
	Type string `json:"type,omitempty"`
}

type metaReply struct {
	Status *dbstat.FrontendStatus `json:"status,omitempty"`
	Blocks *int64                 `json:"blocks,omitempty"`
}

type selectorStatusReply struct {
	Nodes []workspace.Node `json:"nodes,omitempty"`
}

type addWorkspaceSelectorReply struct {
	Nodes []workspace.Node `json:"nodes,omitempty"`
}

type selectorResultsReply struct {
	Results []workspace.TransactionWithTimestamp `json:"transactions,omitempty"`
}

type shortestTransactionPathReply struct {
	Msg          string                   `json:"msg,omitempty"`
	Transactions []db.FrontendTransaction `json:"transactions"`
}

type fingerprintScore struct {
	Score        float64 `json:"score"`
	Txhash       string  `json:"txhash"`
	SessionCount int     `json:"session_count"`
}

type spendingFingerprintReply struct {
	Msg               string             `json:"msg,omitempty"`
	SessionCount      int                `json:"session_count"`
	FingerprintScores []fingerprintScore `json:"fingerprint_scores"`
}

type addWorkspaceNodesReply struct {
	Nodes            []workspace.Node `json:"nodes,omitempty"`
	DuplicateNodeUID string           `json:"duplicateNodeUID,omitempty"`
	// ClusterTooLarge is true if one of the supplied nodes is a cluster with too many outputs
	ClusterTooLarge bool `json:"clusterTooLarge,omitempty"`
}
type addWorkspaceNoteReply struct {
	Nodes []workspace.Node `json:"nodes,omitempty"`
}

type deleteWorkspaceNodeReply struct {
	DeletedNodeUIDs []string `json:"deletedNodeUIDs,omitempty"`
}

type workspaceConnectionReply struct {
	AmountTransactions   []db.AmountTransaction   `json:"amountTransactions,omitempty"`
	FrontendTransactions []db.FrontendTransaction `json:"frontendTransactions,omitempty"`
}

type workspacesReply struct {
	Workspaces []workspace.FrontendWorkspace `json:"workspaces,omitempty"`
}

type getWorkspaceReply struct {
	Workspace *workspace.FrontendWorkspace `json:"workspace,omitempty"`
	// Contains all available heuristic descriptors, which define the heuristic interface
	Descriptors []heuristics.Descriptor `json:"descriptors,omitempty"`
}

type getWorkspaceStateReply struct {
	State string `json:"state,omitempty"`
}

type TokenUser struct {
	ID string `json:"uid,omitempty"`
}

// ExtractTokenUser extracts a TokenUser from the context.
func ExtractTokenUser(ctx context.Context) (t TokenUser, err error) {
	userInfo := ctx.Value(MiddlewareContextUser)
	if userInfo == nil {
		err = serror.FromStr("could not extract token user from context")
		return
	}

	tUser, ok := userInfo.(TokenUser)
	if !ok || len(tUser.ID) == 0 {
		err = serror.FromStr("invalid user extracted from context")
		return
	}

	t = tUser

	return
}

type clusterLookupReply struct {
	Clusters []clustering.FrontendCluster `json:"clusters"`
}

type mixingActivityReply struct {
	Msg        string                     `json:"msg,omitempty"`
	Activities []analytics.MixingActivity `json:"activities,omitempty"`
}

type clusterOverviewReply struct {
	Msg      string                           `json:"msg,omitempty"`
	Clusters []clustering.FrontendUserCluster `json:"clusters"`
}

type msgReply struct {
	Msg string `json:"msg"`
}

type blockReply struct {
	Block *db.FrontendBlock `json:"block"`
}

type transactionReply struct {
	Transactions []db.FrontendTransaction `json:"transactions"`
}

type addressReply struct {
	Address *db.FrontendAddress `json:"address"`
}
