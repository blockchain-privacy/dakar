// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package clustering

import (
	"context"
	"errors"
	"gitlab.com/blockchain-privacy/dakar/cmd/cliutil"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics/clustering"
	"gitlab.com/blockchain-privacy/dakar/external"
	"time"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

type ExternalClusterItem struct {
	ClusterID   string
	AddressHash string
}

var (
	ErrTooManyAddresses   = errors.New("request contains more than 1000 addresses")
	ErrShallowCluster     = errors.New("cluster is too small")
	ErrNonExistentAddress = errors.New("address does not exist")
)

// ImportCluster writes the given address relations into the database
func ImportCluster(ctx context.Context, dgraph external.Database, clusters []ExternalClusterItem, userID string) error {
	if userID == "" {
		return serror.FromStr("user ID is not set")
	}

	addrToUID, err := validateAddresses(ctx, dgraph, clusters)
	if err != nil {
		return err
	}

	dbClusters := buildDatabaseClusters(clusters, userID, addrToUID)

	return clustering.AddCustomClusters(ctx, dgraph, dbClusters)
}

// buildDatabaseClusters creates from the given cluster and address to
// uid mapping (hashToUID) custom clusters which can be inserted into the database.
// hashToUID must contain mappings for all addresses contained in clusters.
func buildDatabaseClusters(clusters []ExternalClusterItem, userID string,
	hashToUID map[string]string) []clustering.CustomCluster {
	set := buildClusterSet(clusters)

	clusterTimestamp := time.Now().UTC().Format(time.RFC3339)

	dbClusters := make([]clustering.CustomCluster, 0, len(set))
	for _, c := range set {
		dbCluster := clustering.CustomCluster{
			Type:         clustering.TypeCustom,
			Timestamp:    clusterTimestamp,
			AddressCount: new(len(c)),
			User:         db.UIDNode{UID: userID},
		}

		dbCluster.SetDType()

		for a := range c {
			dbCluster.Addresses = append(dbCluster.Addresses, db.UIDNode{UID: hashToUID[a]})
		}

		dbClusters = append(dbClusters, dbCluster)
	}

	return dbClusters
}

// buildClusterSet returns a cluster to addresses map
func buildClusterSet(clusters []ExternalClusterItem) map[string]map[string]bool {
	set := map[string]map[string]bool{}

	for _, c := range clusters {
		if set[c.ClusterID] == nil {
			set[c.ClusterID] = map[string]bool{}
		}

		set[c.ClusterID][c.AddressHash] = true
	}

	return set
}

// validateAddresses returns an error, if the given cluster items are not valid.
// Returns ErrTooManyAddresses if there are more than 1000 addresses.
// Returns ErrShallowCluster if there are clusters with less than 2 addresses.
// If an address does not exist in the db, an error containing the address hash is returned.
// Returns a mapping from address hash to db UID, if no errors occurred.
func validateAddresses(ctx context.Context, dgraph external.Database,
	clusters []ExternalClusterItem) (map[string]string, error) {
	addresses := make(map[string]bool, len(clusters))
	for _, c := range clusters {
		addresses[c.AddressHash] = true
	}

	// check maximum number of addresses
	if len(addresses) > 1000 {
		return nil, serror.New(ErrTooManyAddresses)
	}

	// check if clusters contain at least two addresses
	clusterSet := buildClusterSet(clusters)
	for _, v := range clusterSet {
		if len(v) < 2 {
			return nil, serror.New(ErrShallowCluster)
		}
	}

	// check if all addresses exist
	dbAddresses, err := db.GetAddressUIDs(ctx, dgraph, cliutil.GetMapKeys(addresses))
	if err != nil {
		return nil, err
	}

	// check if there is some mismatch
	if len(addresses) != len(dbAddresses) {
		for _, a := range dbAddresses {
			delete(addresses, a.Hash)
		}

		return nil, serror.NewWithContext(ErrNonExistentAddress, "address", cliutil.GetOneKey(addresses))
	}

	// build mapping
	hashToUID := map[string]string{}
	for _, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return nil, serror.FromFormat("address invalid: %v", dbAddress)
		}
		hashToUID[dbAddress.Hash] = dbAddress.UID
	}

	return hashToUID, nil
}
