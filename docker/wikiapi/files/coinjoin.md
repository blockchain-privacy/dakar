# CoinJoin

A CoinJoin happens if a transaction is created (signed) by multiple 
users and the transaction outputs can not be clearly linked to the transaction inputs. 
In most CoinJoin systems, this kind of transaction is created multiple times in succession 
to further obscure the ownership of the funds.

See the [Dash Mixing Process page](dash/mixingProcess) for more details on how CoinJoins work in Dash.

The transaction shown below has the usual structure of a CoinJoin:

 - equal number of inputs and outputs
 - each output and input belongs to a different address
 - usually only one amount denomination (in this case 1 BTC)

![coinjoin](img/coinjoin.png)
