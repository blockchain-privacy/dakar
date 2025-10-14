# Dash Mixing Process

Via the Dash mixing process, the ownership of funds can be obscured.
This is done by using funds as inputs for transactions which have shared ownership. The ownership of the resulting
outputs is therefore only known to the respective mixing participant.

The mixing service is provided by the Dash master nodes as part of the Dash blockchain protocol.

## Main Process Steps

- Funds are first split into predefined [denominations](denominations.md) via
[origin transactions](originTransaction.md)
- The prepared denominations are afterward spent by [mixing transactions](mixingTransaction.md)
- The resulting outputs of mixing transactions are either used as inputs for the next mixing transaction in the mixing graph, or
they are spent via a [destination transaction](destinationTransaction.md)

In the Dash network users are continuously mixing coins, resulting in a large interconnected mixing graph.

## Collateral Payment

To increase the privacy of the mixing transaction outputs, the transaction fee is set to zero.
Because of this, there is potential for abuse. To combat this, collaterals have to be paid by the users of the service via a separate transaction type. 
The collateral amount is fixed at 0.001. Users who have to pay a collateral are randomly selected.
