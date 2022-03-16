package output

import (
	"log"
)

// NumDenominations is the number of Dash PrivateSend denominations existing
const NumDenominations = 5

const (
	// MinCollateral is 1/10 of the smallest denomination: round(100001/10).
	MinCollateral = 10000

	// OldMinCollateral is the minimum collateral before the 5th denomination
	// was added in protocol version 70213 it was round(1000010/10): 100000
	// OldMinCollateral = 100000

	// MaxCollateral is the maximum allowed collateral
	MaxCollateral = 40000 // 4*MinCollateral
	// OldMaxCollateral is to old collateral
	OldMaxCollateral = 400000 // 4*OldMinCollateral
)

var denominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

// CountOutputDenominations returns for each denomination how often it occurred in the given outputs
func CountOutputDenominations(outputs []Output) [NumDenominations]int {
	var amounts []int64

	for _, o := range outputs {
		if o.Amount == nil {
			log.Println("error amount not set")
			return [NumDenominations]int{}
		}
		amounts = append(amounts, *o.Amount)
	}

	return CountAmountDenominations(amounts)
}

// CountAmountDenominations returns the number of occurrences of each denomination in the given amounts
func CountAmountDenominations(amounts []int64) (denominations [NumDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range denominationsTypes {
			if amt == v {
				denominations[i]++
				break inner
			}
		}
	}

	return
}
