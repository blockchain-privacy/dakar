// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/external"
	"context"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// ImportAddressExclusions writes the given address relations into the database
func ImportAddressExclusions(ctx context.Context, dgraph external.Database, exclusions []string, userID string) error {
	if userID == "" {
		return serror.FromStr("user ID is not set")
	}

	if len(exclusions) == 0 {
		return serror.FromStr("address exclusion list is empty")
	}

	uids, err := validateExclusionAddresses(ctx, dgraph, exclusions)
	if err != nil {
		return err
	}

	dbExclusions := buildDatabaseAddressExclusions(uids, userID)

	return exclusion.AddAddressExclusions(ctx, dgraph, dbExclusions)
}

func buildDatabaseAddressExclusions(exclusions []string, userID string) exclusion.User {
	dbExclusions := make([]db.UIDNode, len(exclusions))
	for i, e := range exclusions {
		dbExclusions[i] = db.UIDNode{UID: e}
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
func validateExclusionAddresses(ctx context.Context, dgraph external.Database, exclusions []string) ([]string, error) {
	// check maximum number of items
	if len(exclusions) > 10000 {
		return nil, serror.New(ErrTooManyAddresses)
	}

	if len(exclusions) == 0 {
		return nil, serror.FromStr("empty argument")
	}

	addresses := map[string]bool{}
	for _, c := range exclusions {
		addresses[c] = true
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

	uids := make([]string, len(dbAddresses))
	for i, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return nil, serror.FromFormat("address invalid: %v", dbAddress)
		}
		uids[i] = dbAddress.UID
	}

	return uids, nil
}
