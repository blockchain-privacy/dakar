// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/user"
	"gitlab.com/blockchain-privacy/dakar/db/workspace"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func TestGetAndRefreshWorkspace(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx := t.Context()

	// create dgraph user and workspace for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	wsUID1, err := workspace.AddWorkspace(ctx, dbHandle, "test1", userUID, "")
	require.NoError(t, err)

	m := NewMutex()
	tests := []struct {
		name              string
		workspaceUID      string
		wantWorkspaceName string
		wantErr           bool
	}{
		{
			name:              "invalid workspace UID and name",
			workspaceUID:      "",
			wantWorkspaceName: "",
			wantErr:           true,
		},
		{
			name:              "existing workspace",
			workspaceUID:      wsUID1,
			wantWorkspaceName: "test1",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := GetAndRefreshWorkspace(ctx, dbHandle, m, tt.workspaceUID, userUID)
			if tt.wantErr {
				require.Error(t, err, "name", tt.name)
			} else {
				require.NoError(t, err, "name", tt.name)
				require.Equal(t, tt.wantWorkspaceName, ws.Name, "name", tt.name)
			}
		})
	}
}

const userCount = 3
const workspacesPerUser = 5

func getUserAndWorkspaces(t *testing.T, dbHandle external.Database) ([]string, [][]string, [][]string) {
	ctx := t.Context()
	m := NewMutex()
	users := make([]string, userCount)
	userToWorkspaces := make([][]string, userCount)
	workspaceToNodes := make([][]string, workspacesPerUser)
	for i := range userCount {
		userUID, err := user.CreateNewUser(ctx, dbHandle)
		require.NoError(t, err)
		userToWorkspaces[i] = make([]string, workspacesPerUser)
		users[i] = userUID
		for y := range workspacesPerUser {
			userToWorkspaces[i][y], err = workspace.AddWorkspace(ctx, dbHandle, "test1", userUID, "")
			require.NoError(t, err)

			txUids := []string{"82c973129dc13f84f137c0958c3a6ee875fdb066957abb7fd797ae8845c8689d",
				"040c7a2b65f2f5130f49e244cc8dfcd306bc2873ea34d0a4933d07c73293c536",
				"af25e5385300cfbec9ecba1e7c75035b1c1e77853250db08ac7e455476f5c310",
				"2c101a43418eb86a3a0be0485de90da51dfbf4732ae5c057408cce73aa6f816e"}
			newNodes := make([]workspace.Node, len(txUids))
			for x, query := range txUids {
				newNode, err := workspace.SearchForNode(ctx, dbHandle, query)
				require.NoError(t, err)

				newNodes[x] = *newNode
			}

			addedNodes, _, err := AddNodes(ctx, dbHandle, m, userToWorkspaces[i][y], userUID, newNodes)
			require.NoError(t, err)
			workspaceToNodes[y] = make([]string, len(txUids))
			for x, addedNode := range addedNodes {
				workspaceToNodes[y][x] = addedNode.UID
			}
		}
	}

	return users, userToWorkspaces, workspaceToNodes
}

func TestGetAndRefreshWorkspaceParallel(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)
	users, userToWorkspaces, _ := getUserAndWorkspaces(t, dbHandle)
	ctx := t.Context()
	m := NewMutex()
	wg := sync.WaitGroup{}
	const numCalls = 100
	errChan := make(chan error, numCalls)

	for range numCalls {
		userIndex := rand.IntN(userCount)              //nolint:gosec
		workspaceIndex := rand.IntN(workspacesPerUser) //nolint:gosec
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws, err := GetAndRefreshWorkspace(ctx, dbHandle, m, userToWorkspaces[userIndex][workspaceIndex], users[userIndex])
			if err != nil {
				errChan <- err
			} else if ws == nil {
				errChan <- serror.FromStr("received nil workspace")
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		require.NoError(t, err)
	}
}

func TestWorkspaceUsageParallel(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	users, userToWorkspaces, workspaceToNodes := getUserAndWorkspaces(t, dbHandle)
	ctx := t.Context()
	m := NewMutex()

	txUIDs := []string{"e3902f922a4d5de38b5fd287e52f39d7c41edddcefd91c5c8347dceb3e4c25a9",
		"d0bc5aba5a81df73b706d7819956fb298e03baf52a97c736bb588dfd3586e849",
		"e5a76d4f80bd03f378fc40b550ea8fde9ca6b1dd0916a15c7f22f30947bbe896"}

	newNodes := make([]workspace.Node, len(txUIDs))
	newNodeUIDs := make([]string, len(txUIDs))
	for x, query := range txUIDs {
		newNode, err := workspace.SearchForNode(ctx, dbHandle, query)
		require.NoError(t, err)

		newNodes[x] = *newNode
		newNodeUIDs[x] = newNode.UID
	}

	// 100 * 3
	errChan := make(chan error, 300)

	wg := sync.WaitGroup{}
	for range 100 {
		userIndex := rand.IntN(userCount)              //nolint:gosec
		workspaceIndex := rand.IntN(workspacesPerUser) //nolint:gosec
		userUID := users[userIndex]
		workspaceUID := userToWorkspaces[userIndex][workspaceIndex]

		// move nodes
		wg.Add(1)
		go func() {
			defer wg.Done()

			nodes := make([]workspace.Node, len(workspaceToNodes[workspaceIndex]))
			for i, n := range workspaceToNodes[workspaceIndex] {
				nodes[i] = workspace.Node{
					UID: n,
					X:   new(float32(3)),
					Y:   new(float32(4)),
				}
			}

			err := UpdateNodeCoordinates(ctx, dbHandle, m, workspaceUID, userUID, nodes)
			if err != nil {
				errChan <- err
			}
		}()

		// add nodes
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, _, err := AddNodes(ctx, dbHandle, m, workspaceUID, userUID, newNodes)
			if err != nil {
				errChan <- err
			}
		}()

		// delete nodes
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := DeleteNodes(ctx, dbHandle, m, workspaceUID, userUID, newNodeUIDs)
			if err != nil && !errors.Is(err, errNodeNotFound) {
				if err != nil {
					errChan <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		require.NoError(t, err)
	}
}

func TestDeleteNodes(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)
	users, userToWorkspaces, workspaceToNodes := getUserAndWorkspaces(t, dbHandle)
	m := NewMutex()
	u := users[0]
	ws := userToWorkspaces[0][0]
	tx1 := workspaceToNodes[0][0]
	tx2 := workspaceToNodes[0][1]
	tx3 := workspaceToNodes[0][2]

	// create selector option
	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	require.NoError(t, err)

	opt := workspace.TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: new(int64(1))},
		InputRange:  &workspace.AmountRange{Min: new(int64(1000000)), Max: new(int64(10000000))},
		OutputRange: &workspace.AmountRange{Min: new(int64(1)), Max: new(int64(10000000))},
	}

	optJSON, err := json.Marshal(opt)
	require.NoError(t, err)

	selector1, err := workspace.InsertSelector(t.Context(), dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: string(optJSON),
		Parent:  &db.UIDNode{UID: tx1},
	}, u, ws)
	require.NoError(t, err)

	selector2, err := workspace.InsertSelector(t.Context(), dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: string(optJSON),
		Parent:  &db.UIDNode{UID: tx2},
	}, u, ws)
	require.NoError(t, err)

	selector3, err := workspace.InsertSelector(t.Context(), dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: string(optJSON),
		Parent:  &db.UIDNode{UID: tx3},
	}, u, ws)
	require.NoError(t, err)

	selector4, err := workspace.InsertSelector(t.Context(), dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: string(optJSON),
		Parent:  &db.UIDNode{UID: selector3},
	}, u, ws)
	require.NoError(t, err)

	// tx3 -> selector3 -> selector4 -> selector5
	selector5, err := workspace.InsertSelector(t.Context(), dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: string(optJSON),
		Parent:  &db.UIDNode{UID: selector4},
	}, u, ws)
	require.NoError(t, err)

	// call this so workspace state contains the selector
	w, err := GetAndRefreshWorkspace(t.Context(), dbHandle, m, ws, u)
	require.NoError(t, err)
	require.NotEmpty(t, w.Nodes)

	// deleting transaction should delete its child selector
	nodes, err := DeleteNodes(t.Context(), dbHandle, m, ws, u, []string{tx1})
	require.NoError(t, err)
	require.Contains(t, nodes, selector1)
	require.Contains(t, nodes, tx1)

	// deleting selector first and then tx
	nodes, err = DeleteNodes(t.Context(), dbHandle, m, ws, u, []string{selector2, tx2})
	require.NoError(t, err)
	require.Contains(t, nodes, selector2)
	require.Contains(t, nodes, tx2)

	// deleting partial selector chain and transaction afterward
	nodes, err = DeleteNodes(t.Context(), dbHandle, m, ws, u, []string{selector4, tx3})
	require.NoError(t, err)
	require.Contains(t, nodes, tx3)
	require.Contains(t, nodes, selector3)
	require.Contains(t, nodes, selector4)
	require.Contains(t, nodes, selector5)

	// deleting an already deleted selector
	_, err = DeleteNodes(t.Context(), dbHandle, m, ws, u, []string{selector5})
	require.Error(t, err)
}
