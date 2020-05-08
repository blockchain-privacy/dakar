# DashRPC

## Info

This is a Dash transaction processor, written in Go.

## Dependencies

* btcsuite - core wire layer
* Badger - high-level data storage and partially processed blockchain data

## Start
* Setup `dashd` and let it sync. A GUI is available via `dash-qt`. Dash can be downloaded [here](https://www.dash.org/downloads/). Verify the file hashes.
* Download submodules
```bash
git submodule update --init --recursive
```
* Build the `crawler`
```bash
go build ./cmd/crawler.go
```
* Launch the Dash daemon `dashd` with RPC user and password. In this example the default values from [crawler.go](cmd/crawler.go) are used.
```bash
dashd -rpcuser=rpc1user -rpcpassword=1234pass
```
* Execute the benchmark. You should get output like in the benchmarking section.
```bash
./crawler -benchmark
```

## Benchmarking

* Mariusz's laptop (Intel i7, 7th gen, 3.3GHz, SSD)
```
Elapsed time: 2m4.847019277s
Performance: 499 ms/block
```

* Michaels's computer (Intel Xeon E3-1230v3, 3.3Ghz, SSD)
```
Elapsed time: 1m28.169110927s
Performance: 352 ms/block
```
