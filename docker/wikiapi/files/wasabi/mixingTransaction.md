# Wasabi 2.0 Mixing Transaction

Wasabi 2.0 mixing transactions use outputs of other Wasabi 2.0 mixing transactions or [Wasabi 2.0 origin transactions](wasabi/originTransaction.md)
and create new outputs, which are either mixed again or spent by a Wasabi 2.0 destination transaction.

## Classification Criteria

- all output scripts must be unique
- input amounts must be at least 5000 satoshi
- at least half of the output amounts must be part of the defined [denominations](wasabi/denominations.md)
- the number of outputs must be at least the number of minimum participants (number of inputs divided by 10)
- transactions must contain at least one uncommon denomination

## Example Transaction

`244ce74b5581727721adfbda0f36298e4f36bbff6fa7555f37215e698a35c204`
