package constants

var validDashTransactionTypes = map[string]bool{TypeDashOrigin: true, TypeDashMixing: true,
	TypeDashDestination: true, TypeDashCC: true, TypeDashCP: true}

var validWasabi2TransactionTypes = map[string]bool{TypeWasabi2Origin: true, TypeWasabi2Mixing: true,
	TypeWasabi2Destination: true}

const (
	TypeDashOrigin      = "origin"
	TypeDashMixing      = "mixing"
	TypeDashDestination = "destination"
	TypeDashCC          = "collateral creation"
	TypeDashCP          = "collateral payment"

	TypeWasabi2Origin      = "wasabi origin"
	TypeWasabi2Mixing      = "wasabi mixing"
	TypeWasabi2Destination = "wasabi destination"
)

// IsValidDashTransactionType returns true if the provided string maps to a dash transaction type
func IsValidDashTransactionType(t string) bool {
	return validDashTransactionTypes[t]
}

// IsValidWasabi2TransactionType returns true if the provided string maps to a wasabi 2.0 transaction type
func IsValidWasabi2TransactionType(t string) bool {
	return validWasabi2TransactionTypes[t]
}
