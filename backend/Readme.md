# Backend

This is the backend of Dakar. It crawls the Dash blockchain and exposes its data via a REST API.

## Dependencies

* `btcsuite` - blockchain rpc client access
* `Dgraph` - data storage and processed blockchain data
* `grpc` - network communication
* `ristretto` - in-memory cache for API requests
* `gonum` - graph algorithms
* `prometheus client` - metrics

For a more detailed overview check [here](./go.mod).

## Development

Guide

1. Coding must be done through feature branches
1. Work must be linked to issues from the issue tracker
1. Work should be documented 
1. Work should have unit tests associated, when appropriate  
1. New work should undergo code-review before merging  
1. Small editorial and documentation work can be done directly in `master`
1. [Propogate](https://dave.cheney.net/2015/11/05/lets-talk-about-logging) and [wrap](https://blog.golang.org/go1.13-errors) errors. 
In short: Propogate errors with additional information up to the `main` package and log them there. Do not log errors in other package than `main`. 
Only log if there is an error. Do not log metrics.

Branches
* `master` - main stable dev branch, must compile and should work.
*  feature branches - main mechanism for new work.

## Start
### Setup Dash
* Setup `dashd` and let it sync. A GUI is available via `dash-qt`. Dash can be downloaded [here](https://www.dash.org/downloads/). Verify the file hashes.
* Launch the Dash daemon `dashd` with RPC user and password. In this example the default values from [crawler.go](cmd/crawler.go) are used.
```shell script
dashd -rpcuser=rpc1user -rpcpassword=1234pass
```

### Setup Dgraph
* Change to the `docker` directory and create a new external docker network
```shell script
cd <project_dir>/backend/docker
docker network create dgraph_default
```
* Change the whitelisted ip range in `docker-compose.yml` to your private ip (line 29)
* Execute `docker-compose up` to start Dgraph
* After the startup is complete the database explorer `Ratel` is available via `http://localhost:8000/?local`

### Setup Crawler
* Build the dash version of the crawler
```shell script
cd <project_dir>/backend/cmd/crawler
make dash
```

* Launch the crawler with the following command
```shell script
# -reset will delete all data of the dgraph instance and setup a new schema
./crawler_dash -reset
```
* The REST API can be accessed via the address printed in the standard output.
Check the [crawler description](cmd/crawler/Readme.md) for more details. 
Example output:

```commandline
Dakar v1.0.0
crawler 2021/01/04 11:40:37 main.go:31: Dash mode active
server  2021/01/04 11:40:37 server.go:19: Starting server at endpoint http://localhost:8081
process 2021/01/04 11:40:38 processor.go:31: [Starting crawling at Id: 17940, Hash: 000000000171e06d339fdb33e02eb61ab63415e079a43481bd7cb7b852c4cf4b]
```
### Setup Frontend

* Switch to frontend folder `cd <project_dir>/app`
* Upgrade dependencies `yarn upgrade`
* Start dev server `yarn serve`

Example output:
```text
App running at:
- Local:   http://localhost:8082/ 
- Network: http://<your-private-ip>:8082/
```
