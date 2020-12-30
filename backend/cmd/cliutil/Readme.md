# CLI-Util

This is a module to centralize CLI flags parsing and error handling for Dakar command line utilities. 
Additionally, included are helper functions.

## Using this module

In the Go code import the module and specify via the constants provided by the module the flags which should be used. 

```go
package main

import (
	cli "backend/cmd/cliutil"
)

cliArgs, err := cli.BuildArgs(cli.RpcUser, cli.RpcPassword)
	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}
```

## Adding a new flag

Add the new flag `newFlag` to the "enum". Exported Variables must me uppercase.

```go
const (
	RpcUser Flag = iota
    RpcPassword
	NewFlag
	...
)
```

Add it to the return type `Arguments`.

```go 
type Arguments struct {
	RpcUser     string
    RpcPassword string
	NewFlag     int
	...
}
```

Create a function which sets up the flag. 

Conventions
- Add the default value to the flag description
- The flag should be completely lowercase
- For boolean flags default value is `false`

```go
func addRpcUser(v *string) {
	flag.StringVar(v, "rpcuser", "rpc1user", "Dash RPC user (default: rpc1user)")
}

func addNewFlag(v *int) {
	flag.IntVar(v, "newflag", 0, "New flag description (default: 0)")
}
```
Connect it all in the function `BuildArgs` by adding the new flag to the switch block in the for loop.

```go
for _, f := range flags {
    switch f {
    case RpcUser:
        addRpcUser(&args.RpcUser)
        break
    case NewFlag:
        addNewFlag(&args.NewFlag)
        break
    ...
    }
} 
```

If the new flag needs some **simple** input verification, implement it in this module. Additionally, add the new flag to the table of this `Readme` file.

## Available CLI flags

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
| txsearch | < empty string > | Last PrivateSend transaction hash (default: none) |
| txinfo | < empty string > | Get information about the given transaction hash (default: none) |
| addrcluster | < empty string > | Create cluster for the given address (default: none) |
| btc | false | Switch to Bitcoin mode (default: false) |
