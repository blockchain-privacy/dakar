# CoinJoin Heuristics

CoinJoin heuristics are analytical tools used to identify potential senders and receivers involved in CoinJoin transactions, employing distinct methodologies to narrow down the pool of possible candidates and enhancing the accuracy of transaction analysis. By applying these heuristics, users can gain insights into the relationships and interactions within CoinJoin transactions, ultimately improving their understanding of privacy-enhancing techniques in blockchain networks.

## Types

Depending on the transaction type, some of the following heuristic types are available.

### Lookup Direction: Reverse

#### Wasabi 2.0

- Denomination type
- One source by time
- One source by depth
- Perfect match
- Reverse amount
- Reverse lookup by time
- Reverse lookup by depth

### Lookup Direction: Forward

- Forward amount
- Forward lookup

## Modifiers

The behavior of each heuristic can be modified by the following options:

- Use custom clusters: Use predefined defined custom clusters in combination with multi-input clusters when executing the heuristic
- Use address exclusion list: Do not traverse outputs belonging to the predefined address exclusion list
- Exclude spending gaps: Do not traverse output which have a spending gap
