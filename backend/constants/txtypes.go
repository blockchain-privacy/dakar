package constants

var validTransactionTypes = map[string]bool{TypeDashOrigin: true, TypeDashMixing: true,
	TypeDashDestination: true, TypeDashCC: true, TypeDashCP: true, TypeWasabi2Origin: true, TypeWasabi2Mixing: true,
	TypeWasabi2Destination: true, TypeWhirlpoolOrigin: true, TypeWhirlpoolMixing: true,
	TypeWhirlpoolDestination: true}

const (
	TypeDashOrigin      = "origin"
	TypeDashMixing      = "mixing"
	TypeDashDestination = "destination"
	TypeDashCC          = "collateral creation"
	TypeDashCP          = "collateral payment"

	TypeWasabi2Origin      = "wasabi 2.0 origin"
	TypeWasabi2Mixing      = "wasabi 2.0 mixing"
	TypeWasabi2Destination = "wasabi 2.0 destination"

	TypeWhirlpoolOrigin      = "whirlpool origin"
	TypeWhirlpoolMixing      = "whirlpool mixing"
	TypeWhirlpoolDestination = "whirlpool destination"

	// AllMixingTypes is a helper for query construction
	AllMixingTypes = "\"" + TypeWasabi2Mixing + "\",\"" +
		TypeWhirlpoolMixing + "\",\"" + TypeDashMixing + "\""
	// AllDestinationTypes is a helper for query construction
	AllDestinationTypes = "\"" + TypeWasabi2Destination + "\",\"" +
		TypeWhirlpoolDestination + "\",\"" + TypeDashDestination + "\""
)

// IsTransactionType returns true if the provided string is equal to a transaction type
func IsTransactionType(t string) bool {
	return validTransactionTypes[t]
}

// IsMixingTransaction returns true for mixing types
func IsMixingTransaction(transactionType string) bool {
	return transactionType == TypeDashMixing || transactionType == TypeWasabi2Mixing || transactionType == TypeWhirlpoolMixing
}
