#!/bin/sh

# runs the stack including bitcoind, btc dakar and btc dgraph
docker compose --env-file .env.local -f docker-compose.yml -f compose-btc.yml "$@"
