# Custom Clusters

Custom clusters allow merging multi-input address clusters. The resulting clusters can be used when creating CoinJoin heuristics.

## Example

From multi-input address clustering we have

- cluster A containing address_1, address_2 and address_3
- cluster B containing address_4, address_5 and address_6

By creating a custom cluster containing address_2 and address_4, cluster A and cluster B will be merged into a super cluster. 


## Import

Import address clusters by uploading a CSV-file. The file must have two columns, where the first column contains an identifier for each cluster and the second column the addresses. The file may contain at maximum 1 000 clusters.

The example below would create two clusters with 2 addresses per cluster.

```text
cluster-id,addresshash
1,address_2
1,address_4
2,address_10
2,address_11
```
