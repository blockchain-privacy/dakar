# CoinJoin Heuristics

CoinJoin heuristics are analytical tools used to identify potential senders and receivers involved in CoinJoin transactions, employing distinct methodologies to narrow down the pool of possible candidates and enhancing the accuracy of transaction analysis. By applying these heuristics, users can gain insights into the relationships and interactions within CoinJoin transactions, ultimately improving their understanding of privacy-enhancing techniques in blockchain networks.

## Types

Depending on the transaction type, different kinds of CoinJoin heuristics are available.

### Lookup Direction: Reverse

#### Wasabi 2.0 and Whirlpool

- One source by time
- One source by depth
- Reverse lookup by time
- Reverse lookup by depth
- Reverse amount

#### Dash

- Denomination type
- One source by time
- Perfect match
- Reverse amount
- Reverse lookup by time

### Lookup Direction: Forward

#### Wasabi 2.0 and Whirlpool

- Forward lookup by time
- Forward lookup by depth

#### Dash

- Forward lookup by time
- Forward amount

## Modifiers

The behavior of each heuristic can be modified by the following options:

- Use custom clusters: Use a predefined list of custom clusters in combination with multi-input clusters when executing the heuristic
- Exclude spending gaps: Do not traverse outputs which have a spending gap
