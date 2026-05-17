# Destination Transaction Fingerprinting

Destination transactions spend outputs of mixing transactions. Often these mixing transactions have been created in different time frames, sometimes multiple days or weeks apart. This is due to the user mixing funds at different times.

A destination transaction spending outputs from multiple time frames makes it unique. The more time frames it is connected to the more unique it is. As users often don't spend all their mixed funds via a single destination transaction, it is possible to find other destination transactions of the same user if they spend from the same time frames. 

Destination transaction fingerprinting compares the connected time frames of the analyzed destination transaction with all other destination transactions. Its result are destination transactions which spend outputs from similar time frames.


