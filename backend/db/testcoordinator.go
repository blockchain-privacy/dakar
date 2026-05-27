// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"log"
	"sync"
	"testing"

	"gitlab.com/blockchain-privacy/dakar/external"
)

type TestCoordinator struct {
	dbConnection external.Database
	dbHostname   string
	dbUser       string
	dbPassword   string
}

var singletonCoordinator *TestCoordinator
var once sync.Once

// getTestCoordinator returns a singleton TestCoordinator with the database, mutex and hostname filled
func getTestCoordinator() *TestCoordinator {
	once.Do(func() {
		dbName, ok := GetDBName()
		if !ok {
			log.Fatal("environment variable " + EnvDBHostname + " is not set")
		}

		user := GetDBUser()
		passwd := GetDBPassword()

		ctx, cancel := GetShortTaskContext()
		defer cancel()

		graphDB, err := external.CreateClientWithNamespace(ctx, dbName+":9080", user, passwd, 0)
		if err != nil {
			log.Panic(err)
			return
		}

		singletonCoordinator = &TestCoordinator{dbConnection: graphDB, dbHostname: dbName,
			dbUser: user, dbPassword: passwd}
	})

	return singletonCoordinator
}

// GetDBConnectionWithOptions returns a database connection to a new namespace.
// If setContent is true, the database schema will be set and filled based on fileKey.
// If fileKey is empty, a database connection with no data will be returned.
func GetDBConnectionWithOptions(t *testing.T, setContent bool, fileKey string) external.Database {
	t.Helper()

	if !DoDBTests() {
		t.SkipNow()
		return nil
	}

	c := getTestCoordinator()

	// if no reusable namespace is available, then we need to create new namespace
	// create dgraph client
	nsID, err := c.dbConnection.CreateNamespace(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	graphDB, err := external.CreateClientWithNamespace(t.Context(), c.dbHostname+":9080",
		c.dbUser, c.dbPassword, nsID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		graphDB.Close()

		ctx, cancel := GetTaskContext()
		defer cancel()
		if err := c.dbConnection.DropNamespace(ctx, nsID); err != nil {
			t.Fatal(err)
		}
	})

	if setContent {
		ChangeDBContent(graphDB, fileKey)
	}

	return graphDB
}

// GetDBConnection returns a database connection to a new namespace.
// If fileKey is empty, a database connection with no data will be returned.
func GetDBConnection(t *testing.T, fileKey string) external.Database {
	return GetDBConnectionWithOptions(t, true, fileKey)
}

// GetBareDBConnection returns a database connection with no data and no schema set.
func GetBareDBConnection(t *testing.T) external.Database {
	return GetDBConnectionWithOptions(t, false, "")
}

func ChangeDBContent(dbHandle external.Database, fileKey string) {
	var fileBytes []byte

	switch fileKey {
	case UseClassifierFile:
		fileBytes = ClassifierFile
	case UseBlockFile:
		fileBytes = BlockFile
	case UsePrivacyFile:
		fileBytes = PrivacyFile
	case UseBTCPrivacyFile:
		fileBytes = BTCPrivacyFile
	case "":
	default:
		log.Panic("invalid file key")
	}

	err := WithRetry(func() error {
		return SetupSchema(dbHandle)
	}, retrySleepDuration)
	if err != nil {
		log.Panic("could not set up schema", err)
	}

	if fileBytes != nil {
		ctx, cancel := GetTaskContext()
		defer cancel()
		if err := InsertArbitraryJSON(ctx, dbHandle, fileBytes); err != nil {
			log.Panic("could not upsert block data", err)
		}
	}
}
