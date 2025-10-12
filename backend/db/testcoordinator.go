package db

import (
	"backend/external"
	"log"
	"sync"
	"testing"
)

type TestCoordinator struct {
	dbConnection external.Database
	dbHostname   string
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

		ctx, cancel := GetShortTaskContext()
		defer cancel()

		graphDB, err := external.CreateClient(ctx, dbName+":9080", 0)
		if err != nil {
			log.Panic(err)
			return
		}

		if !external.WaitForDatabase(graphDB) {
			log.Panic("Could not connect to database", err)
			return
		}

		singletonCoordinator = &TestCoordinator{dbConnection: graphDB, dbHostname: dbName}
	})

	return singletonCoordinator
}

// GetDBConnectionWithOptions returns a database connection to a new namespace.
// If setContent is true, the database schema will be set and filled based on fileKey.
// If fileKey is empty, a database connection with no data will be returned.
func GetDBConnectionWithOptions(t *testing.T, setContent bool, fileKey string) external.Database {
	t.Helper()

	if !doDBTests() {
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

	graphDB, err := external.CreateClient(t.Context(), c.dbHostname+":9080", nsID)
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

	if !external.WaitForDatabase(graphDB) {
		t.Fatal("Could not connect to database", err)
	}

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

	if err := SetupSchema(dbHandle); err != nil {
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
