# Wasabi 2.0 Denominations

Wasabi 2.0 uses denominations to increase the privacy of its CoinJoin transactions. 
Inputs of Wasabi 2.0 mixing transactions are split into denominations and change outputs. 

## Denomination List
Wasabi 2.0 defines the following denominations: `5000, 6561, 8192, 10000, 13122, 16384, 19683, 20000, 32768, 39366, 50000, 59049, 65536, 100000, 118098, 131072, 177147, 200000, 262144, 354294, 500000, 524288, 531441, 1000000, 1048576, 1062882, 1594323, 2000000, 2097152, 3188646, 4194304, 4782969, 5000000, 8388608, 9565938, 10000000, 14348907, 16777216, 20000000, 28697814, 33554432, 43046721, 50000000, 67108864, 86093442, 100000000, 129140163, 134217728, 200000000, 258280326, 268435456, 387420489, 500000000, 536870912, 774840978, 1000000000, 1073741824, 1162261467, 2000000000, 2147483648, 2324522934, 3486784401, 4294967296, 5000000000, 6973568802, 8589934592, 10000000000, 10460353203, 17179869184, 20000000000, 20920706406, 31381059609, 34359738368, 50000000000, 62762119218, 68719476736, 94143178827, 100000000000, 137438953472`.

## Common and Uncommon Denominations

The denominations used by Wasabi 2.0 include commonly used amounts used in non-CoinJoin transactions, such as `200000` or `100000000`. We define these denominations (multiples of `5000`) as *common* denominations and all other denominations (not multiples of `5000`) as *uncommon* denominations.

The differentiation between common and uncommon denominations helps with classifying Wasabi 2.0 transactions and to decide when to show the denomination highlight banner.

## Documentation

Read more about this topic in the [Wasabi 2.0 documentation](https://docs.wasabiwallet.io/FAQ/FAQ-UseWasabi.html#what-are-the-equal-denominations-created-in-a-coinjoin-round).
