package status

import (
	"backend/db"
	"backend/external"
	"errors"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

var dbHandle external.Database

const dockerName = "dgraphtest_status"
const alphaPort = "20002"

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:         dockerName,
		Repository:   "dgraph/standalone",
		Tag:          "v21.03.2",
		ExposedPorts: []string{"9080"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9080": {{HostIP: "0.0.0.0", HostPort: alphaPort}},
		},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostName := os.Getenv("YOUR_APP_DB_HOST")

	// create dgraph client
	graphDB, c, err := db.CreateClient(hostName + ":" + alphaPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	dbHandle = graphDB

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		if db.IsConnectionEstablished(graphDB) {
			return nil
		}

		return errors.New("database not ready yet")
	}); err != nil {
		if err = c.Close(); err != nil {
			fmt.Println(err)
		}

		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Panicf("Could not purge resource: %s", err)
		}
		log.Panicf("Could not connect to database: %s", err)
	}

	code := m.Run()

	if err = c.Close(); err != nil {
		fmt.Println(err)
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Panicf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestGetCrawlerStatus(t *testing.T) {
	// crawler status not yet set
	_, err := GetCrawlerStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(dbHandle, true))

	status, err := GetCrawlerStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsCrawling)

	// set not crawling
	require.NoError(t, SetCrawling(dbHandle, false))

	status, err = GetCrawlerStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsCrawling)
}

func TestGetClassifierStatus(t *testing.T) {
	// classifier status not yet set
	_, err := GetClassifierStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set classifying
	require.NoError(t, SetClassifying(dbHandle, true))

	status, err := GetClassifierStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClassifying)

	// set not classifying
	require.NoError(t, SetClassifying(dbHandle, false))

	status, err = GetClassifierStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClassifying)
}

func TestGetClusteringHMIStatus(t *testing.T) {
	// clustering status not yet set
	_, err := GetClusteringHMIStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringHMI(dbHandle, true))

	status, err := GetClusteringHMIStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringHMI(dbHandle, false))

	status, err = GetClusteringHMIStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetClusteringFMIStatus(t *testing.T) {
	// clustering status not yet set
	_, err := GetClusteringFMIStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringFMI(dbHandle, true))

	status, err := GetClusteringFMIStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringFMI(dbHandle, false))

	status, err = GetClusteringFMIStatus(dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetHighestBlockID(t *testing.T) {
	// nothing set yet -> should fail
	_, err := GetHighestBlockID(dbHandle)
	require.Error(t, err)
}

func TestGetFrontendStatus(t *testing.T) {
	require.NoError(t, db.DropAll(dbHandle))

	// nothing set yet -> should fail
	_, err := GetFrontendStatus(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(dbHandle, true))
	require.NoError(t, SetLastBlockID(dbHandle, 50))

	status, err := GetFrontendStatus(dbHandle)
	require.NoError(t, err)
	require.True(t, status.IsCrawling)
	require.Equal(t, status.LastBlockID, uint64(50))
}

func TestGetMeta(t *testing.T) {
	// nothing set yet -> should fail
	_, err := GetMeta(dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set schema version
	require.NoError(t, InitializeMeta(dbHandle, "Dash"))

	metaResult, err := GetMeta(dbHandle)
	require.NoError(t, err)
	require.NotNil(t, metaResult.SchemaVersion)
	require.Equal(t, *metaResult.SchemaVersion, db.SchemaVersion)
	require.Equal(t, metaResult.BlockchainMode, "Dash")
	require.NotEmpty(t, metaResult.CreationTime)
}
