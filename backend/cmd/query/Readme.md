# Query

Perform queries on the Dgraph DB. Possibles query types are:

- chartdir: Create various charts of stored data

Example:

```bash
./query -logfile /tmp/query.log -chartdir /tmp/output
```

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| logfile | < empty string > | Specify log file (default: none) |
| dbhost | 0.0.0.0 | Dgraph host IP (default: 0.0.0.0) |
| dbport | 9080 | Dgraph port (default: 9080) |
| chartdir | < empty string > | Output directory for charts (default: none) |
