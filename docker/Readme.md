# Usage

## Only Dgraph and Dash node

To startup Dgraph and dash node, simply run

```
docker-compose --env-file .env.local up
```

`docker-compose.yml` will be picked up and this contains setting up docker with dgraph, ratel, and dashd
Your local environmental variables should be declared in .env.local


## Dgraph, dashd and front-end for development

```
docker-compose --env-file .env.local -f docker-compose.yml -f docker-compose.dev.yml up
```

For convenience, the dev docker setup command is run through `docker-compose-all-local.sh` bash script.
