# Crawler

This is the blockchain crawler. It loads data from `dashd` and stores it in its own database.

## Examples

Write to a log file, reset the database and start crawling continuously at block height 1.
```sh
./crawler -continuous -logfile /tmp/crawler.log -reset
```

Write to a log file, reset the database and start crawling from block height 1268019 to 1269019
```sh
./crawler -start 1268019 -stop 1269019 -logfile /home/dark/crawler.log -reset
```

Print the current status of the database
```sh
./crawler -status
```

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| continuous | false | Continuously syncs the whole chain (default: false) |
| reset | false | Remove all data from the database (default: false) |
| rpcuser | rpc1user | Dash RPC user (default: rpc1user) |
| rpcpassword | 1234pass | Dash RPC password (default: 1234pass) |
| start | 0 | Start Block Id (default: 0)|
| stop | 0 | Stop Block Id (default: 0) |
| status | false | Prints current processing status (default: false) |
| benchmark | false | Run short performance test (default: false) |
| rpchost | 0.0.0.0 | Dash RPC host IP (default: 0.0.0.0) |
| rpcport | 9998 | Dash RPC port (default: 9998) |
| logfile | < empty string > | Specify log file (default: none) |
