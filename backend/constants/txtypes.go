package constants

var validTransactionTypes = map[string]bool{TypeOrigin: true, TypeMixing: true,
	TypeDestination: true, TypeCC: true, TypeCP: true}

const (
	TypeOrigin      = "origin"
	TypeMixing      = "mixing"
	TypeDestination = "destination"
	TypeCC          = "collateral creation"
	TypeCP          = "collateral payment"
)

// IsValidTransactionType returns true if the provided string maps to a transaction type
func IsValidTransactionType(t string) bool {
	return validTransactionTypes[t]
}
