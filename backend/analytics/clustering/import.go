package clustering

import (
	"backend/cmd/cliutil"
	"backend/db/address"
	"backend/db/analytics/clustering"
	"backend/external"
	"errors"
	"fmt"
)

type ExternalClusterItem struct {
	ClusterID   string
	AddressHash string
}

var (
	ErrTooManyAddresses = errors.New("request contains more than 1000 addresses")
	ErrShallowCluster   = errors.New("cluster is too small")
)

// ImportCluster writes the given address relations into the database
func ImportCluster(dgraph external.Database, clusters []ExternalClusterItem, userID string) error {
	if userID == "" {
		return errors.New("user ID is not set")
	}

	dbClusters := buildDatabaseClusters(clusters, userID)
	if err := clustering.AddCustomClusters(dgraph, dbClusters); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func buildDatabaseClusters(clusters []ExternalClusterItem, userID string) []clustering.CustomCluster {
	set := buildClusterSet(clusters)
	var dbClusters []clustering.CustomCluster
	for _, c := range set {
		numAddresses := len(c)
		dbCluster := clustering.CustomCluster{
			Type:         "custom",
			AddressCount: &numAddresses,
			User:         clustering.HollowUser{Uid: userID},
		}

		dbCluster.SetDType()

		for a := range c {
			dbCluster.Addresses = append(dbCluster.Addresses, clustering.HollowAddress{Uid: a})
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

// ValidateAddresses returns an error is the given cluster items are not valid.
// Returns ErrTooManyAddresses if there are more than 1000 addresses.
// Returns ErrShallowCluster if there are clusters with less than 2 addresses.
// If an address does not exist on the db an error containing the address hash is returned.
func ValidateAddresses(dgraph external.Database, clusters []ExternalClusterItem) error {
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
		return ErrTooManyAddresses
	}

	// check if clusters contain at least two addresses
	clusterSet := buildClusterSet(clusters)
	for _, v := range clusterSet {
		if v == nil || len(v) < 2 {
			return ErrShallowCluster
		}
	}

	// check if all addresses exist
	dbAddresses, err := address.CheckAddressExistence(dgraph, uniqueAddresses)
	if err != nil {
		return err
	}

	// check if there is some mismatch
	if len(addresses) != len(dbAddresses) {
		for _, a := range dbAddresses {
			delete(addresses, a)
		}

		// get one nonexistent address from map
		var nonAddress string
		for k := range addresses {
			nonAddress = k
			break
		}

		return errors.New("address '" + nonAddress + "' does not exist")
	}

	return nil
}
