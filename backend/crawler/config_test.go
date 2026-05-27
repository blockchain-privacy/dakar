// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewBitcoinConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewBitcoinConfig())
}

func TestNewDashConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewDashConfig())
}

func TestNewDogecoinConfig(t *testing.T) {
	require.NotEqual(t, Config{}, NewDogecoinConfig())
}

func Test_preprocessConfig(t *testing.T) {
	require.NotEqual(t, Config{}, preprocessConfig(Config{
		BlockchainName:       "some_blockchain_name",
		BlockTime:            time.Hour,
		NewBlockIntervalTime: 0,
		ForkRangeLimit:       0,
		PubKeyHashAddrID:     0,
	}))
}
