# CoinJoin

A CoinJoin happens if a transaction is created (signed) by multiple 
users and the transaction outputs can not be clearly linked to the transaction inputs. 
In most CoinJoin systems, this kind of transaction is created multiple times in succession 
to further obscure the ownership of the funds.

Depending on the CoinJoin implemenation the transaction format follows a common structure. Possible characteristics are:

- Transaction use a predefined list of amounts (denominations)
- Before the CoinJoin can start, outputs have to be transformed into denominations
- Mixing transactions have a set amount of input and ouputs
- Mixing Fees are payed via a separate transaction type

See the [Dash Mixing Process page](dash/mixingProcess.md) for more details on how CoinJoins work in Dash.

The transaction shown below has the usual structure of a CoinJoin:
