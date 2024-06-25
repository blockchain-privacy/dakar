package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/external"
	"github.com/qrest/gomisc/serror"
)

// ImportAddressExclusions writes the given address relations into the database
func ImportAddressExclusions(dgraph external.Database, exclusions []string, userID string) error {
	if userID == "" {
		return serror.NewStackErrorStr("user ID is not set")
	}

	if len(exclusions) == 0 {
		return serror.NewStackErrorStr("address exclusion list is empty")
	}

	uids, err := validateExclusionAddresses(dgraph, exclusions)
	if err != nil {
		return err
	}

	dbExclusions := buildDatabaseAddressExclusions(uids, userID)

	return exclusion.AddAddressExclusions(dgraph, dbExclusions)
}

func buildDatabaseAddressExclusions(exclusions []string, userID string) exclusion.User {
	dbExclusions := make([]exclusion.AddressExclusions, len(exclusions))
	for i, e := range exclusions {
		dbExclusions[i] = exclusion.AddressExclusions{
			UID: e,
		}
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
func validateExclusionAddresses(dgraph external.Database, exclusions []string) ([]string, error) {
	// check maximum number of items
	if len(exclusions) > 10000 {
		return nil, serror.NewStackError(ErrTooManyAddresses)
	}

	if len(exclusions) == 0 {
		return nil, serror.NewStackErrorStr("empty argument")
	}

	addresses := map[string]bool{}
	for _, c := range exclusions {
		addresses[c] = true
	}

	// check if all addresses exist
	dbAddresses, err := db.GetAddressUIDs(dgraph, cliutil.GetMapKeys(addresses))
	if err != nil {
		return nil, err
	}

	// check if there is some mismatch
	if len(addresses) != len(dbAddresses) {
		for _, a := range dbAddresses {
			delete(addresses, a.Hash)
		}

		return nil, serror.NewStackErrorf("%s: %w", cliutil.GetOneKey(addresses), ErrNonExistentAddress)
	}

	// build mapping

	uids := make([]string, len(dbAddresses))
	for i, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return nil, serror.NewStackErrorf("address invalid: %v", dbAddress)
		}
		uids[i] = dbAddress.UID
	}

	return uids, nil
}
