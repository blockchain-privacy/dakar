package analytics

import "fmt"

type Origin struct {
	Hash           string
	BlockHash      string
	Fee            string
	PrivacyType    string
	BlockId        uint64
	BlockTimestamp string
}

func (o Origin) String() string {
	return fmt.Sprintf("Hash: %s, block hash: %s, fee: %s, privacy type: %s, block id: %d, timestamp: %s",
		o.Hash, o.BlockHash, o.Fee, o.PrivacyType, o.BlockId, o.BlockTimestamp)
}
