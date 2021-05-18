package analytics

type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// IsHeuristicWorkerEnabled controls if analysing is allowed
	IsHeuristicWorkerEnabled bool
	// IsClassifyingEnabled controls if classifying is allowed
	IsClassifyingEnabled bool
	// ClassifierStartBlock is the block id after classifications starts.
	ClassifierStartBlock uint64
}

func NewDashConfig() Config {
	return Config{
		BlockchainName:           "Dash",
		IsHeuristicWorkerEnabled: true,
		// after block height 323756 the first mixing transactions with the
		// most recent format (same number of inputs and outputs) appear
		IsClassifyingEnabled: true,
		ClassifierStartBlock: 323756,
	}
}

func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:           "Bitcoin",
		IsHeuristicWorkerEnabled: false,
		IsClassifyingEnabled:     false,
	}
}

func NewDogecoinConfig() Config {
	return Config{
		BlockchainName:           "Doge",
		IsHeuristicWorkerEnabled: false,
		IsClassifyingEnabled:     false,
	}
}
