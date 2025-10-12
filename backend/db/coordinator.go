package db

import (
	"backend/external"
	"log"
	"sync"
)

type TestCoordinator struct {
	namespacesMutex sync.Mutex
	namespaces      map[uint64]*TestDB
	dbConnection    external.Database
	dbHostname      string
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

		singletonCoordinator = &TestCoordinator{
			namespacesMutex: sync.Mutex{},
			namespaces:      make(map[uint64]*TestDB),
			dbConnection:    graphDB,
			dbHostname:      dbName,
		}
	})

	return singletonCoordinator
}

// GetDBConnection returns a database connection. This may be a connection
// to a newly created db namespace or a reused one.
// If an empty string is passed, a database connection with no data will be returned.
func GetDBConnection(fileKey string) *TestDB {
	c := GetTestCoordinator()
	c.namespacesMutex.Lock()

	for k, n := range c.namespaces {
		if !n.IsDirty.Load() && !n.InUse.Load() && n.FileNameKey == fileKey {
			n.InUse.Store(true)
			c.namespaces[k] = n
			c.namespacesMutex.Unlock()
			info("reusing", fileKey, "namespace", n.NsID)
			return n
		}
	}

	c.namespacesMutex.Unlock()

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

	var newTestDb TestDB
	newTestDb.DB = graphDB
	newTestDb.NsID = nsID
	ChangeDBContent(&newTestDb, fileKey)

	return &newTestDb
}

// GetBareDBConnection returns a database connection with no data and no schema set.
func GetBareDBConnection() *TestDB {
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

	var newTestDb TestDB
	newTestDb.DB = graphDB
	newTestDb.NsID = nsID

	return &newTestDb
}

func ChangeDBContent(dbHandle *TestDB, fileKey string) {
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

	dbHandle.IsDirty.Store(false)
	dbHandle.InUse.Store(true)
	dbHandle.FileNameKey = fileKey
	c := GetTestCoordinator()
	c.namespacesMutex.Lock()
	defer c.namespacesMutex.Unlock()
	c.namespaces[dbHandle.NsID] = dbHandle
}
