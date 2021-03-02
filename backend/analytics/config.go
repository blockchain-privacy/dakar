package analytics

type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// IsAnalysingEnabled controls if analysing is allowed
	IsAnalysingEnabled bool
	// AnalyseStartBlock is the block id after which we start analysing. found empirically.
	AnalyseStartBlock uint64
}

func NewDashConfig() Config {
	return Config{
		BlockchainName:     "Dash",
		IsAnalysingEnabled: true,
		// found empirically.
		AnalyseStartBlock: 206940,
	}
}

func NewBitcoinConfig() Config {
	return Config{
		BlockchainName:     "Bitcoin",
		IsAnalysingEnabled: false,
	}
}

func NewDogecoinConfig() Config {
	return Config{
		BlockchainName:     "Doge",
		IsAnalysingEnabled: false,
	}
}
