# Address Exclusions

Address exclusions allow to exclude outputs linked to addresses from being traversed by CoinJoin heuristics. This is useful if a set of addresses is known to belong to an actor which is not a target of a CoinJoin heuristic. By reducing the amount of outputs being traversed, the number of CoinJoin heuristics results will also be reduced. 

## Requirements

For address exclusions to have a significant impact, they have to prevent the traversal of at least one mixing transaction. As mixing transactions often have a large amount of inputs and outputs, a large amount of address exclusions is required.

Example:

Mixing transaction A is connected to mixing transaction B via 5 outputs. Mixing transaction C is connected to mixing transaction B via 2 outputs. Mixing transaction B is not traversed, only if all 7 outputs (5 from A, 2 from C) are excluded by excluding their 7 respective addresses.


## Import

Address exclusions can be import via a file, which consists of a list of address hashes, separated by new line characters. The file must *not* contain a header. The file may contain at maximum 10 000 addresses.

The example below would add 4 addresses to the exclusion list.

```text
address_1
address_2
address_3
address_4
```
