package analytics

import (
	"backend/cmd/cliutil"
	"backend/db/address"
	"backend/db/analytics/exclusion"
	"backend/external"
	"errors"
	"fmt"
)

// ImportAddressExclusions writes the given address relations into the database
func ImportAddressExclusions(dgraph external.Database, exclusions []string, userID string) error {
	if userID == "" {
		return errors.New("user ID is not set")
	}

	if len(exclusions) == 0 {
		return errors.New("address exclusion list is empty")
	}

	err, uids := validateExclusionAddresses(dgraph, exclusions)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	dbExclusions := buildDatabaseAddressExclusions(uids, userID)
	if err := exclusion.AddAddressExclusions(dgraph, dbExclusions); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func buildDatabaseAddressExclusions(exclusions []string, userID string) exclusion.User {
	var dbExclusions []exclusion.AddressExclusions
	for _, e := range exclusions {
		excl := exclusion.AddressExclusions{
			UID: e,
		}

		dbExclusions = append(dbExclusions, excl)
	}

	return exclusion.User{
		UID:        userID,
		Exclusions: dbExclusions,
	}
}

// validateExclusionAddresses returns an error, if the given address hashes are not valid.
// Returns ErrTooManyAddresses if there are more than 20000 addresses.
// If an address does not exist on the db, an error containing the address hash is returned.
// Returns a list of the address UIDs
func validateExclusionAddresses(dgraph external.Database, exclusions []string) (error, []string) {
	// check maximum number of items
	if len(exclusions) > 10000 {
		return ErrTooManyAddresses, nil
	}

	addresses := map[string]bool{}
	for _, c := range exclusions {
		addresses[c] = true
	}

	var uniqueAddresses []string
	for k := range addresses {
		uniqueAddresses = append(uniqueAddresses, k)
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
	var uids []string
	for _, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return fmt.Errorf("address invalid: %v", dbAddress), nil
		}
		uids = append(uids, dbAddress.UID)
	}

	return nil, uids
}
