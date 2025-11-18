# Dakar

Modules 
 - Crawler: Loads data from `dashd` or `bitcoind` via JSON-RPC and stores it in a Dgraph database.
 - Classifier: Classifies transaction stored in Dgraph regarding their classification type (mixing, origin, destination, ...).
 - Clustering: Clusters addresses of non-mixing transactions via the multi-input clustering heuristic.
 - Heuristics: Allows to apply heuristics on classified transactions in conjunction with clustered addresses.
 - API: HTTP-REST server. 

## Stopping the crawler

Do not kill the crawling process, instead send a termination or interrupt signal. The crawler will then gracefully shutdown.

## Initial Setup

Create a new config file. This will not start Dakar.

```shell script
./dakar -createConfig
```

Start Dakar and confirm the reset dialog.

```shell script
./dakar -reset
```

## Environment Variables

| Variable        | Default Value | Description                                                                                                                              |
|-----------------|:-------------:|------------------------------------------------------------------------------------------------------------------------------------------|
| DEV_GRAPH_LIMIT |     unset     | Limits the number of classified transactions loaded for the in-memory graph. Useful for fast startup of the crawler. Must be an integer. |


## Metrics

Dakar exposes prometheus metrics via `\metrics`. This endpoint is secured via HTTP basic authentication.

## Commandline Arguments

|            Flag | Default Value | Description                                                        |
|----------------:|:-------------:|:-------------------------------------------------------------------|
|           reset |     false     | Remove all data from the database (default: false)                 |
|         version |     false     | Show version information                                           |
|    createConfig |     false     | Creates a default config file (default: false)                     |
| ignoresafeguard |     false     | Ignore the crawling safe guard (default: false)                    |
|   upgradeschema |     false     | Upgrade the database schema to the newest version (default: false) |
|      cpuprofile |   \<empty\>   | Path where the cpu profile should be stored (default: \<empty\>)   |
|          config |  config.yml   | Config file path (default: config.yml)                             |


The crawler registers its activity in the underlying Dgraph database to prevent multiple
crawlers accidentally using the same database at the same time. In case of an unexpected shutdown of the crawler, the safeguard might still be set in the database.
With the `ignoresafeguard` flag the safeguard can be ignored and the crawling be resumed.

## Configuration

The crawler is configured via a configuration file.

### Using the configuration file

Create a new config file with the command below. This will create a new config file named `config.yml`.
```shell script
./dakar -createConfig
```

Start the Dakar with a config file in a custom path.
```shell script
./dakar -config path/to/config/file.yml
```

### Target Iteration Duration

The classifier and clustering module support processing multiple blocks in one iteration. 
The target iteration duration can be set in the configuration file via `targetDuration` in the respective module in multiples of seconds.
Increasing `targetDuration`, increases the relative number of blocks being processed per iteration and therefore also increases the load on the system. 
If `targetDuration` is set to 0, each iteration will only process one block. 
Example:
```yaml
classifier:
    active: true
    targetDuration: 10 # seconds 
```
