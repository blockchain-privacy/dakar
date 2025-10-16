// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/external"
	"context"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"time"
)

type Attribution struct {
	AddressHash string
	Tag         string
	Description string
	Source      string
	Category    string
}

// PublicAttributionImport writes the given attributions into the database and returns the number of insertions.
// Attributions for addresses which do not exist are ignored.
// Attributions for addresses which already have a public attribution are ignored.
func PublicAttributionImport(ctx context.Context, dgraph external.Database, attributions []Attribution) (int, error) {
	if len(attributions) == 0 {
		return 0, serror.FromStr("attribution list is empty")
	}

	addrToUID, err := validateAddresses(ctx, dgraph, attributions, true)
	if err != nil {
		return 0, err
	}

	if len(addrToUID) == 0 {
		return 0, nil
	}

	dbAttributions := buildDatabaseAttributionsIfTheyExist(attributions, addrToUID)
	if len(dbAttributions) == 0 {
		return 0, nil
	}

	return len(dbAttributions), attribution.AddAttributions(ctx, dgraph, dbAttributions)
}

// ImportAttribution writes the given address relations into the database
func ImportAttribution(ctx context.Context, dgraph external.Database, attributions []Attribution, userID string) error {
	if userID == "" {
		return serror.FromStr("user ID is not set")
	}

	if len(attributions) == 0 {
		return serror.FromStr("attribution list is empty")
	}

	addrToUID, err := validateAddresses(ctx, dgraph, attributions, false)
	if err != nil {
		return err
	}

	dbAttributions := buildDatabaseAttributions(attributions, userID, addrToUID)

	return attribution.AddAttributions(ctx, dgraph, dbAttributions)
}

func buildDatabaseAttributions(attributions []Attribution, userID string,
	hashToUID map[string]string) []attribution.Attribution {
	attributionTimestamp := time.Now().UTC().Format(time.RFC3339)

	dbAttributions := make([]attribution.Attribution, len(attributions))

	for i, a := range attributions {
		attr := attribution.Attribution{
			Address:     &db.UIDNode{UID: hashToUID[a.AddressHash]},
			Tag:         a.Tag,
			Description: a.Description,
			Source:      a.Source,
			Category:    a.Category,
			Timestamp:   attributionTimestamp,
			IsPublic:    false,
			User:        &db.UIDNode{UID: userID},
		}

		attr.SetDType()

		dbAttributions[i] = attr
	}

	return dbAttributions
}

func buildDatabaseAttributionsIfTheyExist(attributions []Attribution,
	hashToUID map[string]string) []attribution.Attribution {
	attributionTimestamp := time.Now().UTC().Format(time.RFC3339)

	var dbAttributions []attribution.Attribution //nolint:prealloc
	for _, a := range attributions {
		uid, ok := hashToUID[a.AddressHash]
		if !ok {
			continue
		}
		attr := attribution.Attribution{
			Address:     &db.UIDNode{UID: uid},
			Tag:         a.Tag,
			Description: a.Description,
			Source:      a.Source,
			Category:    a.Category,
			Timestamp:   attributionTimestamp,
			IsPublic:    true,
		}

		attr.SetDType()
		dbAttributions = append(dbAttributions, attr)
	}

	return dbAttributions
}

// validateAddresses returns a mapping from address hash to db UID. An error is returned if
// - the given attribution items are not valid
// - there are more than 1000 items (ErrTooManyAddresses)
// - an address does not exist, an error containing the address hash is returned
// If isPublicAttributionImport is set to true, addresses which do not exist
// or already have a public attribution are ignored.
func validateAddresses(ctx context.Context, dgraph external.Database,
	attributions []Attribution, isPublicAttributionImport bool) (map[string]string, error) {
	// check maximum number of items
	if len(attributions) > 1000 {
		return nil, serror.New(ErrTooManyAddresses)
	}

	if len(attributions) == 0 {
		return nil, serror.FromStr("attribution list is empty")
	}

	addresses := map[string]bool{}
	for _, c := range attributions {
		addresses[c.AddressHash] = true
	}

	var dbAddresses []db.Address
	var err error
	if isPublicAttributionImport {
		dbAddresses, err = db.GetAddressUIDsFilterPublicAttribution(ctx, dgraph, cliutil.GetMapKeys(addresses))
	} else {
		dbAddresses, err = db.GetAddressUIDs(ctx, dgraph, cliutil.GetMapKeys(addresses))
	}
	if err != nil {
		return nil, err
	}

	// check if there is some mismatch
	if !isPublicAttributionImport && len(addresses) != len(dbAddresses) {
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
