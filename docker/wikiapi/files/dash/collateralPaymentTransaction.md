# Collateral Payment Transaction

A collateral payment transaction spends funds created by a
[collateral creation transaction](dash/collateralCreationTransaction.md) or
[mixing transaction](dash/mixingTransaction.md), to pay the collateral required
by the [mixing process](dash/mixingProcess.md).

To prevent abuse of the mixing service (because mixing transactions do not carry a transaction fee), the master nodes randomly select a user who has to pay a collateral in each round.