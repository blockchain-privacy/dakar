# Query

Perform queries on the Dgraph DB. Possibles query types are:

- Address clustering: Create a address cluster for a given address (**not implemented yet**)
- Transaction information: Get information about a given transaction
- Transaction search: Perform a recursive transaction search. The given transaction hash must be the end of a PrivateSend transaction graph.

Example:

```bash
# query for tx info "19bb87c250b8e0d5f6230ee2a85adf00b38bf7f02ae2718a3346170926ec4dc7" 
# and log the output
./query -logfile /tmp/query.log -txinfo 19bb87c250b8e0d5f6230ee2a85adf00b38bf7f02ae2718a3346170926ec4dc7
```

## Commandline Arguments

| Flag | Default Value | Description |
|----------|:-------------:|------:|
| txsearch | < empty string > | Last PrivateSend transaction hash (default: none) |
| txinfo | < empty string > | Get information about the given transaction hash (default: none) |
| addrcluster | < empty string > | Create cluster for the given address (default: none) |
| logfile | < empty string > | Specify log file (default: none) |
| dbhost | 0.0.0.0 | Dgraph host IP (default: 0.0.0.0) |
| dbport | 9080 | Dgraph port (default: 9080) |
