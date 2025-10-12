package db

import (
	"backend/external"
	"log"
	"sync"
)

type TestCoordinator struct {
	dbConnection external.Database
	dbHostname   string
}

var singletonCoordinator *TestCoordinator
var once sync.Once

// GetTestCoordinator returns a singleton TestCoordinator with the database, mutex and hostname filled
func GetTestCoordinator() *TestCoordinator {
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

// GetDBConnection returns a database connection. This may be a connection
// to a newly created db namespace or a reused one.
// If an empty string is passed, a database connection with no data will be returned.
func GetDBConnection(fileKey string) external.Database {
	c := GetTestCoordinator()

	ctx, cancel := GetTaskContext()
	defer cancel()

	// if no reusable namespace is available, then we need to create new namespace
	// create dgraph client
	nsID, err := c.dbConnection.CreateNamespace(ctx)
	if err != nil {
		log.Panic(err)
		return nil
	}

	graphDB, err := external.CreateClient(ctx, c.dbHostname+":9080", nsID)
	if err != nil {
		log.Panic(err)
		return nil
	}

	if !external.WaitForDatabase(graphDB) {
		log.Panic("Could not connect to database", err)
		return nil
	}

	ChangeDBContent(graphDB, fileKey)

	return graphDB
}

// GetBareDBConnection returns a database connection with no data and no schema set.
func GetBareDBConnection() external.Database {
	c := GetTestCoordinator()

	ctx, cancel := GetTaskContext()
	defer cancel()

	// if no reusable namespace is available, then we need to create new namespace
	// create dgraph client
	nsID, err := c.dbConnection.CreateNamespace(ctx)
	if err != nil {
		log.Panic(err)
		return nil
	}

	graphDB, err := external.CreateClient(ctx, c.dbHostname+":9080", nsID)
	if err != nil {
		log.Panic(err)
		return nil
	}

	if !external.WaitForDatabase(graphDB) {
		log.Panic("Could not connect to database", err)
		return nil
	}

	return graphDB
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
