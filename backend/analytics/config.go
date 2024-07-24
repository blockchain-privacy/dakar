package analytics

// Config defines configuration values for a specific blockchain type
type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// IsHeuristicWorkerEnabled controls if analysing is allowed
	IsHeuristicWorkerEnabled bool
	// IsClassifyingEnabled controls if classifying is allowed
	IsClassifyingEnabled bool
	// IsHMIClusteringEnabled controls if clustering is allowed
	IsHMIClusteringEnabled bool
	// IsFMIClusteringEnabled controls if clustering is allowed
	IsFMIClusteringEnabled bool
}

// NewDashConfig returns a Config for Dash
func NewDashConfig() Config {
	return Config{
		BlockchainName:           "Dash",
		IsHeuristicWorkerEnabled: true,
		IsClassifyingEnabled:     true,
		IsHMIClusteringEnabled:   true,
		IsFMIClusteringEnabled:   true,
	}
}

// NewBitcoinConfig returns a Config for Bitcoin
func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:           "Bitcoin",
		IsHeuristicWorkerEnabled: false,
		IsClassifyingEnabled:     false,
		IsHMIClusteringEnabled:   false,
		IsFMIClusteringEnabled:   false,
	}
}
