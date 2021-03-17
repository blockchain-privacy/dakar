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
}

func NewDashConfig() Config {
	return Config{
		BlockchainName:     "Dash",
		IsAnalysingEnabled: true,
		// after block height 200000 the first mixing transactions appear
		// which have the most recent format (same number of inputs and outputs)
		AnalyseStartBlock:    200000,
		IsClassifyingEnabled: true,
		ClassifierStartBlock: 200000,
	}
}

func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:       "Bitcoin",
		IsAnalysingEnabled:   false,
		IsClassifyingEnabled: false,
	}
}

func NewDogecoinConfig() Config {
	return Config{
		BlockchainName:       "Doge",
		IsAnalysingEnabled:   false,
		IsClassifyingEnabled: false,
	}
}
