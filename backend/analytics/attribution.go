package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/db/analytics/attribution"
	"backend/external"
	"time"
)

type Attribution struct {
	AddressHash string
	Tag         string
	Description string
	Source      string
	Category    string
}

// ImportAttribution writes the given address relations into the database
func ImportAttribution(dgraph external.Database, attributions []Attribution, userID string, isPublic bool) error {
	if userID == "" {
		return cliutil.NewStackErrorStr("user ID is not set")
	}

	if len(attributions) == 0 {
		return cliutil.NewStackErrorStr("attribution list is empty")
	}

	addrToUID, err := validateAddresses(dgraph, attributions)
	if err != nil {
		return err
	}

	dbAttributions := buildDatabaseAttributions(attributions, userID, addrToUID, isPublic)

	return attribution.AddAttributions(dgraph, dbAttributions)
}

func buildDatabaseAttributions(attributions []Attribution, userID string, hashToUID map[string]string,
	isPublic bool) []attribution.Attribution {
	attributionTimestamp := time.Now().UTC().Format(time.RFC3339)

	dbAttributions := make([]attribution.Attribution, len(attributions))

	for i, a := range attributions {
		attr := attribution.Attribution{
			Address:     &attribution.HollowAddress{UID: hashToUID[a.AddressHash]},
			Tag:         a.Tag,
			Description: a.Description,
			Source:      a.Source,
			Category:    a.Category,
			Timestamp:   attributionTimestamp,
			IsPublic:    isPublic,
		}

		if !isPublic {
			attr.User = &attribution.HollowUser{UID: userID}
		}

		attr.SetDType()

		dbAttributions[i] = attr
	}

	return dbAttributions
}

// validateAddresses returns an error, if the given attribution items are not valid.
// Returns ErrTooManyAddresses if there are more than 1000 items.
// If an address does not exist on the db, an error containing the address hash is returned.
// Returns a mapping from address hash to db UID.
func validateAddresses(dgraph external.Database, attributions []Attribution) (map[string]string, error) {
	// check maximum number of items
	if len(attributions) > 1000 {
		return nil, cliutil.NewStackError(ErrTooManyAddresses)
	}

	if len(attributions) == 0 {
		return nil, cliutil.NewStackErrorStr("attribution list is empty")
	}

	addresses := map[string]bool{}
	for _, c := range attributions {
		addresses[c.AddressHash] = true
	}

	uniqueAddresses := make([]string, 0, len(addresses))
	for k := range addresses {
		uniqueAddresses = append(uniqueAddresses, k)
	}

	// check if all addresses exist
	dbAddresses, err := db.GetAddressUIDs(dgraph, uniqueAddresses)
	if err != nil {
		return nil, err
	}

	// check if there is some mismatch
	if len(addresses) != len(dbAddresses) {
		for _, a := range dbAddresses {
			delete(addresses, a.Hash)
		}

		return nil, cliutil.NewStackErrorf("%s: %w", cliutil.GetOneKey(addresses), ErrNonExistentAddress)
	}

	// build mapping
	hashToUID := map[string]string{}
	for _, dbAddress := range dbAddresses {
		if dbAddress.Hash == "" || dbAddress.UID == "" {
			return nil, cliutil.NewStackErrorf("address invalid: %v", dbAddress)
		}
		hashToUID[dbAddress.Hash] = dbAddress.UID
	}

	return hashToUID, nil
}
