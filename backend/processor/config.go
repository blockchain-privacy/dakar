// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package processor

import "time"

// Config holds data related to the crawling process of a specific blockchain
type Config struct {
	// BlockchainName is the name of the blockchain
	BlockchainName string
	// BlockTime is the average block time
	BlockTime time.Duration
	// NewBlockIntervalTime is the time interval in which the processor checks if a new block is available
	NewBlockIntervalTime time.Duration
	// ForkRangeLimit is the number of blocks which the RPC client must
	// be ahead of the crawler, for the crawler to include new blocks
	// in the database. This is done so potential chain forks/reordering do not need to be handled
	ForkRangeLimit int64
	// PubKeyHashAddrID is the first byte of a P2PKH address
	PubKeyHashAddrID byte
}

func preprocessConfig(c Config) Config {
	c.NewBlockIntervalTime = c.BlockTime / 3
	return c
}

// NewDashConfig returns the configuration for Dash
func NewDashConfig() Config {
	return preprocessConfig(Config{
		BlockchainName:   "Dash",
		BlockTime:        2*time.Minute + 30*time.Second,
		ForkRangeLimit:   500,
		PubKeyHashAddrID: 0x4c,
	})
}

// NewBitcoinConfig returns the configuration for Bitcoin
func NewBitcoinConfig() Config {
	return preprocessConfig(Config{
		BlockchainName:   "Bitcoin",
		BlockTime:        10 * time.Minute,
		ForkRangeLimit:   125,
		PubKeyHashAddrID: 0x00,
	})
}

// NewDogecoinConfig returns the configuration for Dogecoin
func NewDogecoinConfig() Config {
	return preprocessConfig(Config{
		BlockchainName:   "Doge",
		BlockTime:        1 * time.Minute,
		ForkRangeLimit:   1250,
		PubKeyHashAddrID: 0x1e,
	})
}
