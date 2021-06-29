package analytics

// Config defines configuration values for a specific blockchain type
type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// IsHeuristicWorkerEnabled controls if analysing is allowed
	IsHeuristicWorkerEnabled bool
	// IsClassifyingEnabled controls if classifying is allowed
	IsClassifyingEnabled bool
	// ClassifierStartAfterBlock is the block id after classifications starts.
	ClassifierStartAfterBlock uint64
}

// NewDashConfig returns a Config for Dash
func NewDashConfig() Config {
	return Config{
		BlockchainName:           "Dash",
		IsHeuristicWorkerEnabled: true,
		IsClassifyingEnabled:     true,
		// after block height 323756 the first mixing transactions with the
		// most recent format (same number of inputs and outputs) appear
		ClassifierStartAfterBlock: 0,
	}
}

// NewBitcoinConfig returns a Config for Bitcoin
func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:           "Bitcoin",
		IsHeuristicWorkerEnabled: false,
		IsClassifyingEnabled:     false,
	}
}

// NewDogecoinConfig returns a Config for Dogecoin
func NewDogecoinConfig() Config {
	return Config{
		BlockchainName:           "Doge",
		IsHeuristicWorkerEnabled: false,
		IsClassifyingEnabled:     false,
	}
}
