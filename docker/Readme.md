# Usage

## Prerequisites

   * Copy `.env` file to `.env.local` and modify it to suits your own paths.
   * Copy `dash.conf` file to `<dashd path above>/.dashcore/dash.conf`
   * If you have networking issues in your docker, check if your internal network in docker is in the range `172.17.0.0:172.25.0.255` and if not, modify docker-compose.yml file for allowed range for dgraph DB.


## Only Dgraph and Dash node

To startup Dgraph and dashd daemon, simply run:

```
docker-compose --env-file .env.local up
```

`docker-compose.yml` will be picked up and this contains setting up docker with dgraph, ratel, and dashd
Your local environmental variables should be declared in `.env.local`


## Dgraph, dashd and front-end for development

```
docker-compose --env-file .env.local -f docker-compose.yml -f docker-compose.dev.yml up
```

For convenience, the dev docker setup command is run through `dakar-run-local.sh` bash script. Usage:

   * `./dakar-run-local.sh up`
   * `./dakar-run-local.sh down`
   * `./dakar-run-local.sh build`

Once all the containers are instantiated you can point your browser to `http://localhost` for Dakar front-end.

