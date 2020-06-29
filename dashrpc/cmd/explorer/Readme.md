# Explorer

This is the REST API server. Via the API integrates with the frontend.

## Routes

| Route | Description |
|----------| ------:|
| /api/v1/ | Root |
| /api/v1/tx/ | Transaction details |
| /api/v1/blk/ | Block details |
| /api/v1/address/ | Address details |
| /api/v1/meta/ | Database details |

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| db | /tmp/badger | Badger database location (default: /tmp/badger) |
| rpcuser | rpc1user | Dash RPC user (default: rpc1user) |
| rpcpassword | 1234pass | Dash RPC password (default: 1234pass) |
| rpchost | 0.0.0.0 | Dash RPC host IP (default: 0.0.0.0) |
| rpcport | 9998 | Dash RPC port (default: 9998) |
| logfile | < empty string > | Specify log file (default: none) |
| status | false | Prints current processing status (default: false) |
| serverport | 8081 | Explorer server port (default: 8081) |
