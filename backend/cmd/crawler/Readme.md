# Crawler

This is the blockchain crawler. It loads data from `dashd` and stores it in a Dgraph database. 

## Stopping the crawler

Do not kill the crawling process, instead send a termination or interrupt signal. The crawler will then gracefully shutdown.

## Examples

Create a new config file

```shell script
./crawler -createConfig
```

Confirm the reset dialog and start the crawler.
```shell script
echo yes | ./crawler -reset
```

## Metrics

The crawler exposes prometheus metrics via `\metrics`. This endpoint is secured via HTTP basic authentication.

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| ignoresafeguard | false | Ignore the crawling safe guard (default: false) |
| reset | false | Remove all data from the database (default: false) |
| version | false | Show version information |
| createConfig | false | creates a default config file (default: false) |
| config | config.yml | config file path (default: config.yml) |
