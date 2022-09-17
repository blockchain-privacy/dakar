package db

import (
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

const dockerName = "dgraphtest_db"
const alphaPort = "20000"

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
	graphDB, c, err := CreateClient(hostName + ":" + alphaPort)
	if err != nil {
		info(err)
		return
	}

	dbHandle = graphDB

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		if IsConnectionEstablished(graphDB) {
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
		info(err)
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Panicf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestCreateCommaList(t *testing.T) {
	type testCase struct {
		uids   []string
		result string
	}

	var cases = []testCase{
		{
			uids:   []string{},
			result: "",
		},
		{
			uids:   nil,
			result: "",
		},
		{
			uids:   []string{"123", "456"},
			result: "123,456",
		},
		{
			uids:   []string{"123", ""},
			result: "123,",
		},
	}

	for _, c := range cases {
		require.Equal(t, CreateCommaList(c.uids), c.result)
	}
}

func TestCreateCommaArray(t *testing.T) {
	type testCase struct {
		uids   []string
		result string
	}

	var cases = []testCase{
		{
			uids:   []string{},
			result: "[]",
		},
		{
			uids:   nil,
			result: "[]",
		},
		{
			uids:   []string{"123", "456"},
			result: "[123,456]",
		},
		{
			uids:   []string{"123", ""},
			result: "[123,]",
		},
	}

	for _, c := range cases {
		require.Equal(t, CreateCommaArray(c.uids), c.result)
	}
}

func TestDropAll(t *testing.T) {
	require.NoError(t, DropAll(dbHandle))
}
