package analytics

type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// IsAnalysingEnabled controls if analysing is allowed
	IsAnalysingEnabled bool
	// AnalyseStartBlock is the block id after which analyzing starts.
	AnalyseStartBlock uint64
	// IsClassifyingEnabled controls if classifying is allowed
	IsClassifyingEnabled bool
	// ClassifierStartBlock is the block id after classifications starts.
	ClassifierStartBlock uint64
	// IsInMemoryTransactionGraphEnabled controls if transactions get loaded in memory
	IsInMemoryTransactionGraphEnabled bool
}

func NewDashConfig() Config {
	return Config{
		BlockchainName:     "Dash",
		IsAnalysingEnabled: true,
		// after block height 323756 the first mixing transactions with the
		// most recent format (same number of inputs and outputs) appear
		AnalyseStartBlock:                 323756,
		IsClassifyingEnabled:              true,
		ClassifierStartBlock:              323756,
		IsInMemoryTransactionGraphEnabled: true,
	}
}

func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:                    "Bitcoin",
		IsAnalysingEnabled:                false,
		IsClassifyingEnabled:              false,
		IsInMemoryTransactionGraphEnabled: false,
	}
}

func NewDogecoinConfig() Config {
	return Config{
		BlockchainName:                    "Doge",
		IsAnalysingEnabled:                false,
		IsClassifyingEnabled:              false,
		IsInMemoryTransactionGraphEnabled: false,
	}
}
