package user

import (
	"backend/db"
	"backend/external"
	"errors"
	"flag"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

var dbHandle external.Database

const dockerName = "dgraphtest_user"
const alphaPort = "20001"

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		return
	}
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

func TestGenerateRandomPassword(t *testing.T) {
	pw, err := generateRandomPassword()
	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")
	require.EqualValues(t, len(pw), 22, "got random password with wrong size:")

	pw2, err := generateRandomPassword()
	require.Nil(t, err)
	require.NotEmpty(t, pw, "password is empty")

	require.NotEqual(t, pw, pw2)
}

func TestCreateNewUser(t *testing.T) {
	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
}

func TestGetUsers(t *testing.T) {
	users, err := GetUsers(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestGetUsersWithCredentials(t *testing.T) {
	users, err := GetUsersWithCredentials(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, users)
}

func TestDeleteUser(t *testing.T) {
	// create user
	user, err := CreateNewUser(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	// delete user
	require.NoError(t, DeleteUser(dbHandle, user))

	// try to delete user which does not exist
	require.Error(t, DeleteUser(dbHandle, "some_random_uid_which_does_not_exist"))
}
