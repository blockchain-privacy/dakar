# Crawler

This is the blockchain crawler. It loads data from `dashd` and stores it in its own database.

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| db | /tmp/badger | Badger database location (default: /tmp/badger) |
| continue | false | Continue the previously started DB build process |
| rpcuser | rpc1user | Dash RPC user (default: rpc1user) |
| rpcpassword | 1234pass | Dash RPC password (default: 1234pass) |
| start | 0 | Start Block Id (default: 0)|
| stop | 0 | Stop Block Id (default: 0) |
| status | false | Prints current processing status (default: false) |
| benchmark | false | Run short performance test (default: false) |
| excludeaddresses | false | Exclude addresses from saving into the database (default: false) |
| rpchost | 0.0.0.0 | Dash RPC host IP (default: 0.0.0.0) |
| rpcport | 9998 | Dash RPC port (default: 9998) |
| logfile | < empty string > | Specify log file (default: none) |
