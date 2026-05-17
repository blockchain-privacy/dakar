# CoinJoin

Via a CoinJoin transaction the flow of funds in UTXO-based blockchain systems can be obscured. A CoinJoin happens if a transaction is created (signed) by multiple 
users and each individual transaction output can not be linked to each individual transaction input. In most CoinJoin systems, this kind of transaction is created multiple times in succession to further obscure the ownership of the funds.

Depending on the CoinJoin implementation the transaction format follows a common structure. Possible characteristics are:

- Transactions use a predefined list of amounts (denominations)
- Before a CoinJoin can start, outputs have to be split into certain amount denominations
- Mixing transactions have a set amount of input and outputs
- Mixing fees are paid via a separate transaction type

See the [Dash Mixing Process page](dash/mixingProcess.md) for more details on how CoinJoins work in Dash.
