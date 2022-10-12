package main

import (
	cli "backend/cmd/cliutil"
	"backend/db"
	dbus "backend/db/user"
	"backend/external"
	"context"
	"flag"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"log"
	"net/http"
	"net/http/cookiejar"
)

var thisLogger *log.Logger

func initLogger() {
	thisLogger = log.New(log.Writer(), "\033[0;31mdbup\033[0m\t", log.Flags())
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

type Config struct {
	Logfile string `yaml:"logfile"`
	Host    string `yaml:"host"`
	Port    uint   `yaml:"port"`
}

var defaultConfig = Config{
	Logfile: "",
	Host:    "0.0.0.0",
	Port:    9080,
}

// newKratosClient creates a new kratos client
func newKratosClient(endpoint string) (*ory.APIClient, error) {
	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}

	conf.HTTPClient = &http.Client{Jar: cj}

	return ory.NewAPIClient(conf), nil
}

// Simple utility to browse/lookup the TXs from the database
func main() {
	////// SET FLAGS //////

	defaultConfigName := "config.yml"
	var filePath string
	var createConfigFile bool
	cli.SetConfigFlags(defaultConfigName, &filePath, &createConfigFile)
	flag.Parse()

	////// CONFIGURATION FILE HANDLING //////

	if createConfigFile {
		fmt.Println("Generating configuration file ...")

		err := cli.WriteConfig(defaultConfigName, defaultConfig)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("config file", defaultConfigName, "successfully created")
		return
	}

	var config Config
	if err := cli.ReadConfig(filePath, &config); err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(config.Logfile) > 0 {
		if f, err := cli.GetLogfile(config.Logfile); err == nil {
			defer func() {
				if err = f.Close(); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

	initLogger()

	endpoint, err := cli.BuildEndpoint(config.Host, config.Port)
	if err != nil {
		info(err)
		return
	}

	// create dgraph client
	dgraph, c, err := external.CreateClient(endpoint)
	if err != nil {
		info(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			info(err)
		}
	}()

	isSet, err := db.IsSchemaSet(dgraph)
	if err != nil {
		info(err)
		return
	}

	if !isSet {
		info("Schema is not set")
		return
	}

	kratos, err := newKratosClient("http://localhost:4434")
	if err != nil {
		return
	}

	users, err := dbus.GetUsersWithCredentials(dgraph)
	if err != nil {
		info(err)
		return
	}

	for _, u := range users {
		if u.Email == "" || u.Pwhash == "" {
			info("email or password not set, not processing:", u)
			continue
		}

		var roles []string

		for _, r := range u.Roles {
			roles = append(roles, r.Name)
		}

		if len(roles) == 0 {
			info("no roles found, not processing", u)
			continue
		}

		createError := dbus.CreateKratosUser(context.Background(), u.UID, kratos,
			u.Email, &ory.AdminIdentityImportCredentials{
				Password: &ory.AdminCreateIdentityImportCredentialsPassword{
					Config: &ory.AdminCreateIdentityImportCredentialsPasswordConfig{
						HashedPassword: &u.Pwhash,
					}},
			}, roles)
		if createError != nil {
			info(createError)
			return
		}
	}

	//info("increasing schema version ...")
	//err = status.SetSchemaVersion(dgraph, 2)
	//if err != nil {
	//	info(err)
	//	return
	//}
	//info("increased schema version")
}
