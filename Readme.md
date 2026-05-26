# Dakar

Dakar is a blockchain forensics application, focusing on the analysis of CoinJoin transactions.
It consists of a backend implemented in Go and web app written in Vue 3 with the Vuetify UI framework. 

The backend ingests blockchain data via RPC connections to blockchain clients and performs transaction
classification and address clustering on it.

The web app allows exploring the ingested and transformed data. Additionally, graph based editor enables
viewing relationships between transactions and address clusters. Several CoinJoin analysis modules are
available: heuristics, transaction similarity measure, mixing activity overview and more.

## Get Started

### Set up Blockchain Client

* Setup either `dashd` or `bitcoind`
* Configure the RPC connection and set the details in `config.yml` of Dakar

### Setup Dgraph

* Change to the `docker` directory 
* Change the whitelisted ip range in `docker-compose.yml` to include your private docker container ip
* Set the appropriate values in the .env file
* Execute `docker compose up` to start Dgraph and its auth stack
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
Check the [Dakar description](backend/cmd/dakar/Readme.md) for more details. 

### Docker

To create a docker image containing the Dakar executable execute the script below.
```shell script
make docker
```
The image expects the config file to be mounted to `/data/config.yml`.

### Setup Frontend

* Switch to frontend folder `cd app`
* Start dev server `yarn dev`

## Metrics

Metrics are exposed via `/metrics` on a separate port, which is configurable via the config file.

## Running local tests

Some tests require a connection to a dgraph database or a blockchain RPC client.

The command below runs all tests, which don't require a database or RPC connection.
```shell
make test
```

To run database tests, first set up an empty dgraph instance, preferably via [docker](docker/docker-compose_local-test.yml).
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

The API documentation is built with [swaggo](https://github.com/swaggo/swag) using the annotations in the [api](backend/server/api.go) file.

The following command
- compiles the OpenAPI schema the in [openapi](backend/openapi) directory
- builds the [TypeScript client](backend/openapi/client) 
- and publishes it to the Gitlab registry. 

```shell
make openapi-spec && make openapi-client && make openapi-publish
```

Make sure to have the deployment token set in your `~/.yarnrc.yml`:

```yaml
yarnPath: .yarn/releases/yarn-4.0.0.cjs

npmScopes:
  blockchain:
    npmRegistryServer: "https://<gitlab_host>/api/v4/projects/<project_id>/packages/npm/"
    npmAlwaysAuth: true
    npmAuthToken: "<your-deploy-token>"
```

Format the OpenAPI annotations using:

```shell
make openapi-fmt
```

## Authentication and Authorization

Dakar's API uses the UID stored in the request header field `X-dakar-user`. 
Dakar does not authenticate any request. This should be done by a dedicated identity management system.
We provide configuration for the [ory](https://github.com/ory) auth stack, which is also integrated into the web app.
All configuration files are in dev mode.

### OAuth 2.0 client creation
Create a device authorization client with the following commands. 
Use the client id stored in `$code_client` to set up an oauth client.
```shell
cd docker
code_client=$(sudo docker compose --env-file .env.local exec hydra \
    hydra create client \
    --endpoint http://<hydra_endpoint>:4445 \
    --grant-type authorization_code,refresh_token,urn:ietf:params:oauth:grant-type:device_code \
    --response-type code,id_token \
    --name <client_name> \
    --skip-consent \
    --format json \
    --scope openid --scope offline \
    --audience dakar \
    --token-endpoint-auth-method none)
echo $code_client
```

## Development

1. Don't introduce new dependencies unless discussed with the maintainer
2. [Propagate](https://dave.cheney.net/2015/11/05/lets-talk-about-logging) and [wrap](https://blog.golang.org/go1.13-errors) errors.
    1. Propagate errors with additional information up to the `main` package and log them there. Do not log errors in other package than `main`.
       Only log if there is an error. Do not log metrics.
    2. Wrap all native errors via the [StackError](https://gitlab.com/blockchain-privacy/gomisc/serror) type to enable error tracing.

## Citation Information 

If you use this software in your research or work, please cite it using the following BibTeX entry:

```bibtex
@Article{Ziegler2026,
  author    = {Ziegler, Michael Herbert and Nowostawski, Mariusz and Katt, Basel},
  journal   = {SoftwareX},
  title     = {{Dakar: A CoinJoin forensic software}},
  year      = {2026},
  issn      = {2352-7110},
  month     = feb,
  volume    = {33},
  doi       = {10.1016/j.softx.2026.102523},
  publisher = {Elsevier BV},
}
```
