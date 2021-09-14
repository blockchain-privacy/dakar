# Crawler

This is the blockchain crawler. It loads data from `dashd` and stores it in a Dgraph database. 

An HTTP server can be activated (with `-startserver`), which exposes the database entries via a REST API.

## Stopping the crawler

Do not kill the crawling process, instead send a termination or interrupt signal. The crawler will then gracefully shutdown.

## Examples

Write to a log file, reset the database, start the http server on the default port and start crawling at block height 1.

```shell script
./crawler -logfile /tmp/crawler.log -reset
```

Print the current status of the database
```shell script
./crawler -status
```

Confirm the reset dialog and start crawler, classifier and server.
```shell script
echo yes | ./crawler -reset
```

## Metrics

The crawler exposes prometheus metrics via `\metrics`.

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| ignoresafeguard | false | Ignore the crawling safe guard (default: false) |
| reset | false | Remove all data from the database (default: false) |
| rpcuser | rpc1user | Dash RPC user (default: rpc1user) |
| rpcpassword | 1234pass | Dash RPC password (default: 1234pass) |
| status | false | Prints current processing status (default: false) |
| rpchost | 0.0.0.0 | Dash RPC host IP (default: 0.0.0.0) |
| rpcport | 9998 | Dash RPC port (default: 9998) |
| dbhost | 0.0.0.0 | Dgraph host IP (default: 0.0.0.0) |
| dbport | 9080 | Dgraph port (default: 9080) |
| logfile | < empty string > | Specify log file (default: none) |
| disableserver | false | Disable the http server (default: false) |
| disablecrawler | false | Disable the crawler (default: false) |
| disablehmiclustering | false | Disable hierarchical multi-input clustering (default: false) |
| disablefmiclustering | false | Disable flat multi-input clustering (default: false) |
| disableclassifier | false | Disable the classifier (default: false) |
| disableclustering | false | Disable clustering (default: false) |
| serverport | 8081 | Http server port (default: 8081) |
