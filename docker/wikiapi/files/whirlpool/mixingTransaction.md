# Whirlpool Mixing Transaction

Whirlpool mixing transactions use outputs of other Whirlpool mixing transactions or [Whirlpool origin transactions](whirlpool/originTransaction.md)
and create new outputs, which are either mixed again or spent by a Whirlpool destination transaction.


A transaction must satisfy the following conditions to be classified as a Whirlpool mixing transaction:
- must have at least 5 outputs
- must have at maximum 8 outputs
- must have same amount of inputs as outputs
- all input and outputs must have the same [denomination](whirlpool/denominations.md)
- must spend at least one input from a whirlpool origin transaction

Example transaction: `a94c73ab1a0d9a08f14317b4662b0cb932c308d22211ea30fcfa7df023dadf8e`
