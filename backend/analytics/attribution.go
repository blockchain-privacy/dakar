package analytics

import (
	"backend/cmd/cliutil"
	"backend/db/address"
	"backend/db/analytics/attribution"
	"backend/external"
	"errors"
	"fmt"
	"time"
)

type Attribution struct {
	AddressHash string
	Tag         string
	Description string
	Source      string
	Category    string
}

var (
	errTooManyAddresses   = errors.New("request contains more than 1000 addresses")
	errNonExistentAddress = errors.New("address does not exist")
)

// ImportAttribution writes the given address relations into the database
func ImportAttribution(dgraph external.Database, attributions []Attribution, userID string, isPublic bool) error {
	if userID == "" {
		return errors.New("user ID is not set")
	}

	err, addrToUID := validateAddresses(dgraph, attributions)
	if err != nil {
		return err
	}

	dbAttributions := buildDatabaseAttributions(attributions, userID, addrToUID, isPublic)
	if err := attribution.AddAttributions(dgraph, dbAttributions); err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

func buildDatabaseAttributions(attributions []Attribution, userID string, hashToUID map[string]string,
	isPublic bool) []attribution.Attribution {
	attributionTimestamp := time.Now().UTC().Format(time.RFC3339)

	var dbAttributions []attribution.Attribution
	for _, a := range attributions {
		attr := attribution.Attribution{
			Address:     &attribution.HollowAddress{Uid: hashToUID[a.AddressHash]},
			Tag:         a.Tag,
			Description: a.Description,
			Source:      a.Source,
			Category:    a.Category,
			Timestamp:   attributionTimestamp,
			IsPublic:    isPublic,
		}

		if !isPublic {
			attr.User = &attribution.HollowUser{Uid: userID}
		}

		attr.SetDType()

		dbAttributions = append(dbAttributions, attr)
	}

	return dbAttributions
}

// validateAddresses returns an error, if the given attribution items are not valid.
// Returns errTooManyAddresses if there are more than 1000 items.
// If an address does not exist on the db an error containing the address hash is returned.
// Returns a mapping from address hash to db UID.
func validateAddresses(dgraph external.Database, attributions []Attribution) (error, map[string]string) {
	// check maximum number of items
	if len(attributions) > 1000 {
		return errTooManyAddresses, nil
	}

	addresses := map[string]bool{}
	for _, c := range attributions {
		addresses[c.AddressHash] = true
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

		return fmt.Errorf("%s: %w", nonAddress, errNonExistentAddress), nil
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
