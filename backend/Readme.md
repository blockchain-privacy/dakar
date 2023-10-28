 # Backend

This is the backend of Dakar. It crawls the Dash blockchain and exposes its data via a REST API.

* [Backend](#backend)
   * [Dependencies](#dependencies)
   * [Development](#development)
   * [Start](#start)
      * [Setup Dash](#setup-dash)
      * [Setup Dgraph](#setup-dgraph)
      * [Setup Crawler](#setup-crawler)
      * [Docker](#docker)
      * [Setup Frontend](#setup-frontend)
   * [Metrics](#metrics)
   * [Running local tests](#running-local-tests)
   * [OpenAPI Documentation](#openapi-documentation)

## Dependencies

* `btcsuite` - blockchain rpc client access
* `dgraph` - data storage and processed blockchain data
* `grpc` - network communication
* `ristretto` - in-memory cache for API requests
* `gonum` - graph algorithms
* `prometheus client` - metrics
* `ory kratos` - user authentication and credential management

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
### Setup Dash
* Setup `dashd` and let it sync. A GUI is available via `dash-qt`. Dash can be downloaded [here](https://www.dash.org/downloads/). Verify the file hashes.
* Launch the Dash daemon `dashd` with RPC user and password. In this example the default values from [crawler.go](cmd/crawler/main.go) are used.
```shell script
dashd -rpcuser=rpc1user -rpcpassword=1234pass
```

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

### Setup Crawler
* Build the dash version of the crawler
```shell script
cd <project_dir>/backend/cmd/crawler
make dash
```

* Launch the crawler with the following command
```shell script
# -reset will delete all data of the dgraph instance and setup a new schema
./crawler_dash -reset
```
* The REST API can be accessed via the address printed in the standard output.
Check the [crawler description](cmd/crawler/Readme.md) for more details. 
Example output:

```commandline
Dakar v1.0.0
crawler 2021/01/04 11:40:37 main.go:31: Dash mode active
server  2021/01/04 11:40:37 server.go:19: Starting server at endpoint http://localhost:8081
process 2021/01/04 11:40:38 processor.go:31: [Starting crawling at Id: 17940, Hash: 000000000171e06d339fdb33e02eb61ab63415e079a43481bd7cb7b852c4cf4b]
```

### Docker

To create a docker image containing the crawler executable execute the script below.
```shell script
make docker-dash
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
go test -cover -race ./... 
```

To run database tests, first set up an empty dgraph instance, preferably via [docker](../docker/docker-compose_local-test.yml).
Next, set the `DB_TESTS` and `DB_HOSTNAME` environment variables. `DB_HOSTNAME` should be set to the host which runs the database. The port is expected to be `9080`. 
Set parallelism to 1, so database tests of different modules don't interfere which each other.

```shell
DB_TESTS=1 DB_HOSTNAME=localhost go test -p 1 -cover -race ./... 
```

Set the `DB_TESTS`, `DB_HOSTNAME` and `RPC_TESTS` environment variables to run all tests.

```shell
RPC_TESTS=1 DB_TESTS=1 DB_HOSTNAME=localhost go test -p 1 -cover -race ./... 
```

## OpenAPI Documentation

The API documentation is built with [swaggo](https://github.com/swaggo/swag) using the annotations in the [api](server/api.go) file.

The following command
- compiles the OpenAPI schema the in [openapi](openapi) directory
- builds the [Typescript client](openapi/client) 
- and publishes it to the [Gitlab registry](https://git.gvk.idi.ntnu.no/research/blockchain/dakar/-/packages). 

```shell
make swagger-create
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
make swagger-fmt
```
