package analytics

import "fmt"

type Origin struct {
	Uid            string
	Hash           string
	BlockHash      string
	BlockId        uint64
	BlockTimestamp string
}

func (o Origin) String() string {
	return fmt.Sprintf("Uid: %s, Hash: %s, block hash: %s, block id: %d, timestamp: %s",
		o.Uid, o.Hash, o.BlockHash, o.BlockId, o.BlockTimestamp)
}
