# DashRPC

## Info

This is a Dash transaction processor, written in Go.

## Dependencies

* btcsuite - core wire communication layer
* Dgraph - data storage and processed blockchain data


## Development

Guide

1. Coding must be done through feature branches
1. Work must be linked to issues from the issue tracker
1. Work should be documented 
1. Work should have unit tests associated, when appropriate  
1. New work should undergo code-review before merging  
1. Small editorial and documentation work can be done directly in `master`

Branches
* `master` - main stable dev branch, must compile and should work.
* `production` - deployed branch, no work/commits should be done in here.
*  feature branches - main mechanism for new work.

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

* Michael's computer (Intel Xeon E3-1230v3, 3.3Ghz, SSD)
```
Elapsed time: 1m43.415075837s
Performance: 413 ms/block
```
