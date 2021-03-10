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
		// found empirically.
		AnalyseStartBlock:    206940,
		IsClassifyingEnabled: true,
		ClassifierStartBlock: 206940,
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
