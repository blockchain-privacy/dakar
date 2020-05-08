# Dash

## Info

This is a Dash transaction processor, written in Go.

## Dependencies

* btcsuite - core wire layer
* Badger - high-level data storage and partially processed blockchain data


## Start

* Build the `crawler`

```bash
go build ./cmd/crawler.go
```


* Launch the Dash deamon `dashd`

* Execute the benchmark run

```
./crawler -benchmark
```

## Benchmarking

* Mariusz's laptop (Intel i7, 7th gen, 3.3GHz, SSD)
```
Elapsed time: 2m4.847019277s
Performance: 499 ms/block
```


