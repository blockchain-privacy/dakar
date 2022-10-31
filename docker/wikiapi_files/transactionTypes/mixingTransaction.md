# Mixing Transaction

Mixing transactions use outputs of other mixing transactions or [origin transactions](transactionTypes/originTransaction.md)
and create new outputs, which are either mixed again or spent by a destination transaction.


To qualify as a mixing transaction

- the number of inputs and outputs most be equal
- the transaction must be zero
- all input and output amounts must be part of the defined [denominations](denominations.md)


![mixing transaction example](img/mixing.png)