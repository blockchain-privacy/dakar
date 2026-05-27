// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package blockiterator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_scaleBlocksPerIteration(t *testing.T) {
	tests := []struct {
		target            time.Duration
		iterationDuration time.Duration
		blockCount        int64
		want              int64
	}{
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 1,
			blockCount:        5,
			want:              6,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 6,
			blockCount:        3,
			want:              2,
		},
		// not lower than 1
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 6,
			blockCount:        1,
			want:              1,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 6,
			blockCount:        3,
			want:              2,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 5,
			blockCount:        3,
			want:              3,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 0,
			blockCount:        200,
			want:              200,
		},
		// should add 10%
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 0,
			blockCount:        100,
			want:              110,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 6,
			blockCount:        100,
			want:              75,
		},
		{
			target:            time.Duration(-100),
			iterationDuration: time.Second * 6,
			blockCount:        100,
			want:              1,
		},
		{
			target:            time.Duration(0),
			iterationDuration: time.Second * 6,
			blockCount:        100,
			want:              1,
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, scaleBlocksPerIteration(tt.target, tt.iterationDuration, tt.blockCount))
	}
}
