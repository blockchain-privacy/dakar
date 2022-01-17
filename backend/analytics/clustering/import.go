package clustering

import (
	"backend/cmd/cliutil"
	"backend/db/address"
	"backend/db/analytics/clustering"
	"backend/external"
	"errors"
	"fmt"
	"time"
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
func ImportCluster(dgraph external.Database, clusters []ExternalClusterItem, userID string) error {
	if userID == "" {
		return errors.New("user ID is not set")
	}

	err, addrToUID := validateAddresses(dgraph, clusters)
	if err != nil {
		return err
	}

	dbClusters := buildDatabaseClusters(clusters, userID, addrToUID)
	if err := clustering.AddCustomClusters(dgraph, dbClusters); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func buildDatabaseClusters(clusters []ExternalClusterItem, userID string, hashToUID map[string]string) []clustering.CustomCluster {
	set := buildClusterSet(clusters)

	clusterTimestamp := time.Now().Format(time.RFC3339)

	var dbClusters []clustering.CustomCluster
	for _, c := range set {
		numAddresses := len(c)
		dbCluster := clustering.CustomCluster{
			Type:         "custom",
			Timestamp:    clusterTimestamp,
			AddressCount: &numAddresses,
			User:         clustering.HollowUser{Uid: userID},
		}

		dbCluster.SetDType()

		for a := range c {
			dbCluster.Addresses = append(dbCluster.Addresses, clustering.HollowAddress{Uid: hashToUID[a]})
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
// If an address does not exist on the db an error containing the address hash is returned.
// Returns a mapping from address hash to db UID.
func validateAddresses(dgraph external.Database, clusters []ExternalClusterItem) (error, map[string]string) {
	addresses := map[string]bool{}
	for _, c := range clusters {
		addresses[c.AddressHash] = true
	}

	var uniqueAddresses []string
	for k := range addresses {
		uniqueAddresses = append(uniqueAddresses, k)
	}

	// check maximum number of addresses
	if len(uniqueAddresses) > 1000 {
		return ErrTooManyAddresses, nil
	}

	// check if clusters contain at least two addresses
	clusterSet := buildClusterSet(clusters)
	for _, v := range clusterSet {
		if v == nil || len(v) < 2 {
			return ErrShallowCluster, nil
		}
	}

	// check if all addresses exist
	dbAddresses, err := address.GetAddressUIDs(dgraph, uniqueAddresses)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err), nil
	}

	// check if there is some mismatch
	if len(addresses) != len(dbAddresses) {
		for _, a := range dbAddresses {
			delete(addresses, a.Hash)
		}

		// get one nonexistent address from map
		var nonAddress string
		for k := range addresses {
			nonAddress = k
			break
		}

		return fmt.Errorf("%s: %w", nonAddress, ErrNonExistentAddress), nil
	}

	// build mapping
	hashToUid := map[string]string{}
	for _, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return fmt.Errorf("address invalid: %v", dbAddress), nil
		}
		hashToUid[dbAddress.Hash] = dbAddress.UID
	}

	return nil, hashToUid
}
