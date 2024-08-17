package blockiterator

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_scaleBlocksPerIteration(t *testing.T) {
	tests := []struct {
		target            time.Duration
		iterationDuration time.Duration
		blockCount        uint64
		want              uint64
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
			blockCount:        5,
			want:              2,
		},
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
			want:              1,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 5,
			blockCount:        3,
			want:              4,
		},
		{
			target:            time.Second * 5,
			iterationDuration: time.Second * 0,
			blockCount:        200,
			want:              200,
		},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.want, scaleBlocksPerIteration(tt.target, tt.iterationDuration, tt.blockCount))
	}
}
