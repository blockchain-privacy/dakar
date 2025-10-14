 # Backend

This is the backend of Dakar. It crawls, classifies and clusters either the Dash or Bitcoin blockchain and exposes its data via a RESTful API.

* [Backend](#backend)
   * [Dependencies](#dependencies)
   * [Development](#development)
   * [Start](#start)
      * [Setup Blockchain Client](#setup-blockchain-client)
      * [Setup Dgraph](#setup-dgraph)
      * [Setup Dakar](#setup-dakar)
      * [Docker](#docker)
      * [Setup Frontend](#setup-frontend)
   * [Metrics](#metrics)
   * [Running Local Tests](#running-local-tests)
   * [OpenAPI Documentation](#openapi-documentation)

## Dependencies

* `dgraph` - data storage and processed blockchain data
* `grpc` - network communication
* `ristretto` - in-memory cache for API requests
* `gonum` - graph algorithms
* `prometheus client` - metrics

For a more detailed overview check [here](./go.mod).

## Development

Guide
1. Coding must be done through feature branches
1. Work must be linked to issues from the issue tracker
1. Work should be documented 
1. Work should have unit tests associated, when appropriate  
1. New work should undergo code-review before merging  
1. Small editorial and documentation work can be done directly in `master`
1. [Propagate](https://dave.cheney.net/2015/11/05/lets-talk-about-logging) and [wrap](https://blog.golang.org/go1.13-errors) errors. 
   1. Propagate errors with additional information up to the `main` package and log them there. Do not log errors in other package than `main`. 
   Only log if there is an error. Do not log metrics.
   1. Wrap all native errors via the [StackError](cmd/cliutil/stackerror.go) type to enable error tracing.

Branches
* `master` - main stable dev branch, must compile and should work.
*  feature branches - main mechanism for new work.

## Start
### Setup Blockchain Client
* Setup either `dashd` or `bitcoind`
* Configure the RPC connection and set the details in `config.yml` of Dakar.

### Setup Dgraph
* Change to the `docker` directory and create a new external docker network
```shell script
cd <project_dir>/backend/docker
docker network create dgraph_default
```
* Change the whitelisted ip range in `docker-compose.yml` to your private ip
* Set the appropriate values in the .env file
* Execute `docker-compose up` to start Dgraph
* After the startup is complete the database explorer `Ratel` is available via `http://localhost:8000/?local`

### Setup Dakar
* Build Dakar
```shell script
cd backend
make dakar
cd build
```

* Create a new config file. Change the values in the newly generated `config.yml` to appropriate values.
```shell script
# -createConfig will create a new config file `config.yml` in your current directory
./dakar -createConfig
```

* Launch the Dakar executable with the following command
```shell script
# -reset will delete all data of the dgraph instance and setup a new schema
./build/dakar -reset
```
* The REST API can be accessed via the address printed in the standard output.
Check the [Dakar description](cmd/dakar/Readme.md) for more details. 

### Docker

To create a docker image containing the Dakar executable execute the script below.
```shell script
make docker
```
The image expects the config file to be mounted to `/data/config.yml`.

### Setup Frontend

* Switch to frontend folder `cd <project_dir>/app`
* Upgrade dependencies `yarn upgrade`
* Start dev server `yarn serve`

Example output:
```text
App running at:
- Local:   http://localhost:8082/ 
- Network: http://<your-private-ip>:8082/
```

## Metrics

Metrics are exposed via `/metrics` on a separate port, which is configurable via the config file.

## Running local tests

Some tests require a connection to a dgraph database and a blockchain RPC-client.

The command below runs all tests, which don't require a database and RPC-client.
```shell
make test
```

To run database tests, first set up an empty dgraph instance, preferably via [docker](../docker/docker-compose_local-test.yml).
Set the `DB_TESTS` and `DB_HOSTNAME` environment variables to run database tests. `DB_HOSTNAME` should be set to the host which runs the database. The port is expected to be `9080`.

```shell
export DB_TESTS=1; export DB_HOSTNAME=0.0.0.0; make test
```

Set the `DB_TESTS`, `DB_HOSTNAME`, `RPC_TESTS` and `RPC_HOSTNAME` environment variables to run all tests.

```shell
export DB_TESTS=1; export RPC_TESTS=1; export DB_HOSTNAME=0.0.0.0; export RPC_HOSTNAME=0.0.0.0; make test
```

Additionally, the Dgraph ACL user and password can be configured via `DB_USER` and `DB_PASSWORD`. 

| Environment Variable | Description                               |
|:---------------------|:------------------------------------------|
| DB_TESTS             | Set to enable database tests              |
| DB_HOSTNAME          | The hostname of the dgraph test database  |
| DB_USER              | The ACL user name (default: groot)        |
| DB_PASSWORD          | The ACL password (default: password)      |
| RPC_TESTS            | Set to enable blockchain RPC tests        |
| RPC_HOSTNAME         | The hostname of the blockchain RPC client |


## OpenAPI Documentation

The API documentation is built with [swaggo](https://github.com/swaggo/swag) using the annotations in the [api](server/api.go) file.

The following command
- compiles the OpenAPI schema the in [openapi](openapi) directory
- builds the [Typescript client](openapi/client) 
- and publishes it to the [Gitlab registry](https://git.gvk.idi.ntnu.no/research/blockchain/dakar/-/packages). 

```shell
make openapi-spec && make openapi-client && make openapi-publish
```

Make sure to have the deployment token set in your `~/.yarnrc.yml`:

```yaml
yarnPath: .yarn/releases/yarn-4.0.0.cjs

npmScopes:
  blockchain:
    npmRegistryServer: "https://git.gvk.idi.ntnu.no/api/v4/projects/410/packages/npm/"
    npmAlwaysAuth: true
    npmAuthToken: "<your-deploy-token>"
```

Format the OpenAPI annotations using:

```shell
make openapi-fmt
```
