# Usage

## Stack

- Dgraph: Graph DB for blockchain data
- RPC client: Currently either `dashd` or `bitcoind` which feeds data into Dgraph
- Ory Kratos: User authentication
- Ory Oathkeeper: Request authentication which integrates with Ory Kratos and Ory Keto
- WikiApi: A markdown-to-HTML converter used for displaying documentation on the frontend

## Prerequisites

   * Copy `.env` file to `.env.local` and modify it to suit your preferences
   * Copy `dash.conf` file to `<dashd path above>/.dashcore/dash.conf`
   * Generate signing keys for Ory Oathkeeper: ``docker-compose run oathkeeper credentials generate --alg RS256 > oathkeeper/id_token.jwks.json     ``
   * If networking issues with docker occur, check if your internal network in docker is in the range `172.17.0.0:172.25.0.255` and if not, modify docker-compose.yml file for allowed range for dgraph DB

## Run services

To start the backend services run:

```
docker-compose --env-file .env.local up
```

For convenience, the docker setup command can be run through the `dakar-run-local.sh` script. 

Usage:

   * `./dakar-run-local.sh up`
   * `./dakar-run-local.sh down`
   * `./dakar-run-local.sh build`

Once all the containers are instantiated, start the crawler located in `../cmd/crawler` and the web frontend.
