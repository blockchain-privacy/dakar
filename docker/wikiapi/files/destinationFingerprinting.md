# Destination Transaction Fingerprinting

Destination transactions spend outputs of mixing transactions. Often this mixing transactions have been created in different timeframes, sometimes multiple days or weeks apart. This is due to the user mixing funds at different times.

A destination transaction spending outputs from multiple timeframes makes it unique. The more timeframes it is connected to the more unique it becomes. As users often don't spend all their mixed funds via single destination transaction, it is possible to find other destination transactions of the same user if they spend from the same time frames. 

Destination transaction fingerprinting compares the connected timeframes of the anlayzed destination transaction with all other destination transactions. Its result are destination transactions with similar timeframes.


