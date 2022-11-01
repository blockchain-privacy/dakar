package graph

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_splitTimestampsIntoSessions(t *testing.T) {
	tests := []struct {
		timestamps         []int64
		wantNumberSessions int
	}{
		{
			timestamps:         nil,
			wantNumberSessions: 0,
		},
		{
			timestamps:         []int64{},
			wantNumberSessions: 0,
		},
		{
			timestamps:         []int64{time.Now().Unix()},
			wantNumberSessions: 1,
		},
		{
			timestamps:         []int64{time.Now().Unix(), time.Now().AddDate(0, 0, -5).Unix()},
			wantNumberSessions: 2,
		},
		{
			timestamps: []int64{time.Now().Unix(), time.Now().AddDate(0, 0, -5).Unix(),
				time.Now().AddDate(0, 0, 2).Unix()},
			wantNumberSessions: 3,
		},
		{
			timestamps: []int64{time.Now().Unix(), time.Now().Add(time.Hour * 2).Unix(),
				time.Now().Add(time.Hour * 3).Unix(), time.Now().Add(time.Hour * 4).Unix(),
				time.Now().Add(time.Hour * 5).Unix()},
			wantNumberSessions: 1,
		},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.wantNumberSessions, len(splitTimestampsIntoSessions(tt.timestamps)))
	}
}

func Test_getSessionMeans(t *testing.T) {
	tests := []struct {
		sessions [][]int64
		want     []int64
	}{
		{
			sessions: nil,
			want:     []int64{},
		},
		{
			sessions: [][]int64{{5, 4, 3}, {5, 5}, {0, 3}, {2, 2}},
			want:     []int64{4, 5, 1, 2},
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, getSessionMeans(tt.sessions))
	}
}

func Test_getShortestDistances(t *testing.T) {
	type args struct {
		sessionMeans1 []int64
		sessionMeans2 []int64
	}
	tests := []struct {
		args args
		want []int64
	}{
		{
			args: args{
				sessionMeans1: nil,
				sessionMeans2: nil,
			},
			want: []int64{},
		},
		{
			args: args{
				sessionMeans1: []int64{5, 4, 3, 2},
				sessionMeans2: []int64{5, 4, 3, 2},
			},
			want: []int64{0, 0, 0, 0},
		},
		{
			args: args{
				sessionMeans1: []int64{100, 52},
				sessionMeans2: []int64{105, 50},
			},
			want: []int64{5, 2},
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, getShortestDistances(tt.args.sessionMeans1, tt.args.sessionMeans2))
	}
}
