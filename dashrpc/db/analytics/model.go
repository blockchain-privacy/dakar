package analytics

import "fmt"

type Origin struct {
	Hash           string
	BlockHash      string
	BlockId        uint64
	BlockTimestamp string
}

func (o Origin) String() string {
	return fmt.Sprintf("Hash: %s, block hash: %s, block id: %d, timestamp: %s",
		o.Hash, o.BlockHash, o.BlockId, o.BlockTimestamp)
}
