# Crawler

This is the blockchain crawler. It loads data from `dashd` and stores it in a Dgraph database. 

An HTTP server can be activated (with `-startserver`), which exposes the database entries via a REST API. 

## Crawling Modes

The crawler can be started either in continuous mode (`-continuous`) or in range mode (`-start 1 -stop 10`). 
In range mode some functionality is not completely available. This is the case if outputs of transactions **not** part of the 
range are used as inputs of transactions part of the range. 

## REST API Routes

Routes supported by the REST API. Consume the endpoints via GET requests.

| Route | Description |
|----------| ------:|
| /api/v1/ | Possible routes |
| /api/v1/tx/ | Transaction details |
| /api/v1/blk/ | Block details |
| /api/v1/address/ | Address details |
| /api/v1/meta/ | Database details |
| /api/v1/origins/ | Origins of transaction |

## Stopping the crawler

Do not kill the crawling process, instead send a termination or interrupt signal. The crawler will then gracefully shutdown.

## Examples

Write to a log file, reset the database, start the http server on the default port and start crawling continuously at block height 1.

```shell script
./crawler -continuous -logfile /tmp/crawler.log -reset
```

Write to a log file, reset the database and start crawling from block height 1268019 to 1269019. Also start the http server.
```shell script
./crawler -start 1268019 -stop 1269019 -logfile /home/dark/crawler.log -reset
```

Print the current status of the database
```shell script
./crawler -status
```

Confirm the reset dialog and start crawler, analyzer and server.
```shell script
echo yes | ./crawler -reset -continous
```

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| continuous | false | Continuously syncs the whole chain (default: false) |
| ignoresafeguard | false | Ignore the crawling safe guard (default: false) |
| reset | false | Remove all data from the database (default: false) |
| rpcuser | rpc1user | Dash RPC user (default: rpc1user) |
| rpcpassword | 1234pass | Dash RPC password (default: 1234pass) |
| start | 0 | Start Block Id (default: 0)|
| stop | 0 | Stop Block Id (default: 0) |
| status | false | Prints current processing status (default: false) |
| rpchost | 0.0.0.0 | Dash RPC host IP (default: 0.0.0.0) |
| rpcport | 9998 | Dash RPC port (default: 9998) |
| dbhost | 0.0.0.0 | Dgraph host IP (default: 0.0.0.0) |
| dbport | 9080 | Dgraph port (default: 9080) |
| logfile | < empty string > | Specify log file (default: none) |
| disableserver | false | Disable the http server (default: false) |
| disablecrawler | false | Disable the crawler (default: false) |
| disableanalyzer | false | Disable the analyzer (default: false) |
| serverport | 8081 | Http server port (default: 8081) |
