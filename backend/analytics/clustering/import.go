package clustering

import (
	"backend/db/address"
	"backend/external"
	"errors"
)

type ExternalClusterItem struct {
	ClusterID   string
	AddressHash string
}

var (
	ErrTooManyAddresses = errors.New("request contains more than 1000 addresses")
)

// validateAddresses returns an error is the given addresses are not valid
func validateAddresses(dgraph external.Database, clusters []ExternalClusterItem) error {
	addresses := map[string]bool{}
	for _, c := range clusters {
		addresses[c.AddressHash] = true
	}

	var uniqueAddresses []string
	for k := range addresses {
		uniqueAddresses = append(uniqueAddresses, k)
	}

	if len(uniqueAddresses) > 1000 {
		return ErrTooManyAddresses
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

// ImportCluster writes the given address relations into the database
func ImportCluster(dgraph external.Database, clusters []ExternalClusterItem) error {
	if err := validateAddresses(dgraph, clusters); err != nil {
		return err
	}

	return nil
}
