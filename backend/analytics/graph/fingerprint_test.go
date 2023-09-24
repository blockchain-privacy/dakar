package graph

import (
	"github.com/stretchr/testify/require"
	"sort"
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
		{
			timestamps:         []int64{1531235575, 1531181371, 1531183888, 1531193372, 1531205321, 1530574495, 1531135657},
			wantNumberSessions: 3,
		},
		{
			timestamps: []int64{1531212600, 1531218784, 1531181024, 1531193674, 1531176993, 1531199139, 1531207215,
				1531184567, 1531214700, 1531214043, 1531204878, 1531201698, 1531189758, 1531203632, 1531182759,
				1531195844, 1531202944, 1531210485, 1531178947},
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
		{
			sessions: [][]int64{{1531135657, 1530574495}, {1531181371, 1531183888, 1531193372, 1531205321, 1531235575}},
			want:     []int64{1530855076, 1531199905},
		},
		{
			sessions: [][]int64{{1531212600, 1531218784, 1531181024, 1531193674, 1531176993, 1531199139, 1531207215,
				1531184567, 1531214700, 1531214043, 1531204878, 1531201698, 1531189758, 1531203632, 1531182759,
				1531195844, 1531202944, 1531210485, 1531178947}},
			want: []int64{1531198614},
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

func Test_scoreMeans(t *testing.T) {
	type args struct {
		rootMeans  []int64
		otherMeans []int64
	}
	tests := []struct {
		args    args
		want    float64
		wantErr bool
	}{
		{
			args:    args{},
			want:    0,
			wantErr: true,
		},
		{
			args: args{
				rootMeans:  []int64{1531199726},
				otherMeans: []int64{1530855076, 1531199905},
			},
			want:    2.1176843975392314,
			wantErr: false,
		},
		{
			args: args{
				rootMeans:  []int64{1531199726},
				otherMeans: []int64{1530574495, 1531135657, 1531199905},
			},
			want:    2.1176843975392314,
			wantErr: false,
		},
		{
			args: args{
				rootMeans:  []int64{1531199698},
				otherMeans: []int64{1530574495, 1531135657, 1531199905},
			},
			want:    2.0421204423092894,
			wantErr: false,
		},
		{
			args: args{
				rootMeans: []int64{1531199698},
				otherMeans: []int64{1505288219, 1512635534, 1518554829, 1520623200, 1520856655, 1528519368, 1528686712, 1529408611, 1531199724,
					1533753995, 1533877938, 1534231847, 1534345337, 1535177998, 1535462996},
			},
			want:    3.4303014159363543,
			wantErr: false,
		},
		{
			args: args{
				rootMeans: []int64{1531198614}, // real value
				otherMeans: []int64{1505288219, 1512635534, 1518554829, 1520623200, 1520856655, 1528519368, 1528686712, 1529408611, 1531199724,
					1533753995, 1533877938, 1534231847, 1534345337, 1535177998, 1535462996},
			},
			want:    1.3414487146333909,
			wantErr: false,
		},
		{
			args: args{
				rootMeans: []int64{1531198614}, // real value
				otherMeans: []int64{1505288219, 1512635534, 1518554829, 1520623200, 1520856655, 1528519368,
					1528686712, 1529408611, 1531199724, 1533753995, 1533877938, 1534231847, 1534345336, 1535177998}},
			want:    1.3414487146333909,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := scoreMeans(tt.args.rootMeans, tt.args.otherMeans)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		}
	}
}

func TestCompareTransactions(t *testing.T) {
	tests := []struct {
		timestamps1 []int64
		timestamps2 []int64
		wantScore   float64
	}{
		{
			timestamps1: []int64{1531220838, 1531176993, 1531181024, 1531182759, 1531202944, 1531178947,
				1531193674, 1531218784, 1531184567, 1531189758, 1531195844, 1531199139, 1531203632, 1531204878,
				1531199139, 1531201698, 1531207215, 1531210485, 1531212600, 1531214043, 1531214700},
			timestamps2: []int64{1531235575, 1531181371, 1531183888, 1531193372, 1531205321, 1530574495, 1531135657},
			wantScore:   2.0421204423092894,
		},
		{
			timestamps1: []int64{1531220838, 1531176993, 1531181024, 1531182759, 1531202944, 1531178947,
				1531193674, 1531218784, 1531184567, 1531189758, 1531195844, 1531199139, 1531203632, 1531204878,
				1531199139, 1531201698, 1531207215, 1531210485, 1531212600, 1531214043, 1531214700},
			timestamps2: []int64{1535462996, 1520623200, 1529392931, 1529426674, 1531227936, 1534322685, 1512635534,
				1534355572, 1533753995, 1531128037, 1531207215, 1533753995, 1533881161, 1528686712, 1518554829, 1529406228,
				1531165928, 1531244120, 1533878548, 1534230643, 1505288219, 1534360913, 1533879092, 1534232822, 1534342177,
				1533753995, 1533753995, 1533879092, 1534232076, 1535177998, 1531129913, 1520856655, 1528519368, 1531206804,
				1531235874, 1531251689, 1533872953},
			wantScore: 3.4303014159363543,
		},
		{
			timestamps1: []int64{1531212600, 1531218784, 1531181024, 1531193674, 1531176993, 1531199139, 1531207215,
				1531184567, 1531214700, 1531214043, 1531204878, 1531201698, 1531189758, 1531203632, 1531182759,
				1531195844, 1531202944, 1531210485, 1531178947},
			// 0x3a74eb8
			timestamps2: []int64{1505288219,
				1531207215, 1533872953, 1529406228, 1534232076, 1529426674, 1529392931, 1531235874, 1531227936, 1534230643,
				1534342177, 1531129913, 1535177998, 1534322685, 1512635534, 1528519368, 1528686712, 1531128037, 1531206804,
				1531251689, 1531165928, 1534232822, 1533878548, 1534360913, 1534355572, 1533753995, 1533879092, 1520623200,
				1518554829, 1531244120, 1533881161, 1520856655},
			wantScore: 1.3414487146333909,
		},
		{
			timestamps1: []int64{1531212600, 1531218784, 1531181024, 1531193674, 1531176993, 1531199139, 1531207215,
				1531184567, 1531214700, 1531214043, 1531204878, 1531201698, 1531189758, 1531203632, 1531182759,
				1531195844, 1531202944, 1531210485, 1531178947},
			// 0x34cb589
			timestamps2: []int64{1531181371, 1531183888, 1531193372, 1531205321, 1530574495, 1531135657},
			wantScore:   0.7683578754134952,
		},
	}
	for _, tt := range tests {
		means1 := getSessionMeans(splitTimestampsIntoSessions(tt.timestamps1))
		means2 := getSessionMeans(splitTimestampsIntoSessions(tt.timestamps2))
		score, err := scoreMeans(means1, means2)
		require.NoError(t, err)
		require.EqualValues(t, tt.wantScore, score)
	}
}

// newDestinationGraph creates a new graph with the following destination transaction timestamps:
// - 2020-01-01 10:00 - 11:00, 2020-01-02 15:00 - 17:00 - set 1
// - 2020-01-01 08:00 - 09:00, 2020-01-02 13:00 - 16:00 - set 1
// - 2020-02-01 10:00 - 11:00, 2020-02-02 15:00 - 17:00 - set 2
// - 2020-02-01 08:00 - 09:00, 2020-02-02 13:00 - 16:00 - set 2
// - 2020-03-01 10:00 - 11:00, 2020-03-02 15:00 - 17:00 - set 3
func newDestinationGraph(t *testing.T) *ReversibleGraph {
	graph := NewReversibleGraph(5)

	tx1ts1, err := time.Parse(time.RFC3339, "2020-01-01T10:00:00Z")
	require.NoError(t, err)
	tx1ts2, err := time.Parse(time.RFC3339, "2020-01-01T11:00:00Z")
	require.NoError(t, err)
	tx1ts3, err := time.Parse(time.RFC3339, "2020-01-02T15:00:00Z")
	require.NoError(t, err)
	tx1ts4, err := time.Parse(time.RFC3339, "2020-01-02T17:00:00Z")
	require.NoError(t, err)

	tx2ts1, err := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")
	require.NoError(t, err)
	tx2ts2, err := time.Parse(time.RFC3339, "2020-01-01T09:00:00Z")
	require.NoError(t, err)
	tx2ts3, err := time.Parse(time.RFC3339, "2020-01-02T13:00:00Z")
	require.NoError(t, err)
	tx2ts4, err := time.Parse(time.RFC3339, "2020-01-02T16:00:00Z")
	require.NoError(t, err)

	tx3ts1, err := time.Parse(time.RFC3339, "2020-02-01T10:00:00Z")
	require.NoError(t, err)
	tx3ts2, err := time.Parse(time.RFC3339, "2020-02-01T11:00:00Z")
	require.NoError(t, err)
	tx3ts3, err := time.Parse(time.RFC3339, "2020-02-02T15:00:00Z")
	require.NoError(t, err)
	tx3ts4, err := time.Parse(time.RFC3339, "2020-02-02T17:00:00Z")
	require.NoError(t, err)

	tx4ts1, err := time.Parse(time.RFC3339, "2020-02-01T08:00:00Z")
	require.NoError(t, err)
	tx4ts2, err := time.Parse(time.RFC3339, "2020-02-01T09:00:00Z")
	require.NoError(t, err)
	tx4ts3, err := time.Parse(time.RFC3339, "2020-02-02T13:00:00Z")
	require.NoError(t, err)
	tx4ts4, err := time.Parse(time.RFC3339, "2020-02-02T16:00:00Z")
	require.NoError(t, err)

	tx5ts1, err := time.Parse(time.RFC3339, "2020-03-01T10:00:00Z")
	require.NoError(t, err)
	tx5ts2, err := time.Parse(time.RFC3339, "2020-03-01T11:00:00Z")
	require.NoError(t, err)
	tx5ts3, err := time.Parse(time.RFC3339, "2020-03-02T15:00:00Z")
	require.NoError(t, err)
	tx5ts4, err := time.Parse(time.RFC3339, "2020-03-02T17:00:00Z")
	require.NoError(t, err)

	now := time.Now()

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, PrivacyType: 101},
		TransactionNode{TS: tx1ts1, id: 10, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, PrivacyType: 101},
		TransactionNode{TS: tx1ts2, id: 11, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, PrivacyType: 101},
		TransactionNode{TS: tx1ts3, id: 12, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 1, PrivacyType: 101},
		TransactionNode{TS: tx1ts4, id: 13, PrivacyType: 0}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, PrivacyType: 101},
		TransactionNode{TS: tx2ts1, id: 20, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, PrivacyType: 101},
		TransactionNode{TS: tx2ts2, id: 21, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, PrivacyType: 101},
		TransactionNode{TS: tx2ts3, id: 22, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 2, PrivacyType: 101},
		TransactionNode{TS: tx2ts4, id: 23, PrivacyType: 0}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, PrivacyType: 101},
		TransactionNode{TS: tx3ts1, id: 30, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, PrivacyType: 101},
		TransactionNode{TS: tx3ts2, id: 31, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, PrivacyType: 101},
		TransactionNode{TS: tx3ts3, id: 32, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 3, PrivacyType: 101},
		TransactionNode{TS: tx3ts4, id: 33, PrivacyType: 0}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, PrivacyType: 101},
		TransactionNode{TS: tx4ts1, id: 40, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, PrivacyType: 101},
		TransactionNode{TS: tx4ts2, id: 41, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, PrivacyType: 101},
		TransactionNode{TS: tx4ts3, id: 42, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 4, PrivacyType: 101},
		TransactionNode{TS: tx4ts4, id: 43, PrivacyType: 0}, 0)

	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, PrivacyType: 101},
		TransactionNode{TS: tx5ts1, id: 50, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, PrivacyType: 101},
		TransactionNode{TS: tx5ts2, id: 51, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, PrivacyType: 101},
		TransactionNode{TS: tx5ts3, id: 52, PrivacyType: 0}, 0)
	graph.SetEdgeWithoutOverwrite(
		TransactionNode{TS: now, id: 5, PrivacyType: 101},
		TransactionNode{TS: tx5ts4, id: 53, PrivacyType: 0}, 0)

	return graph
}

func TestSpendingFingerprint(t *testing.T) {
	g := newDestinationGraph(t)

	tests := []struct {
		uid             string
		wantFingerprint []string
		wantNumSessions int
		wantErr         bool
	}{
		{
			uid:             "",
			wantFingerprint: nil,
			wantNumSessions: 0,
			wantErr:         true,
		},
		{
			uid:             ToHex(1),
			wantFingerprint: []string{ToHex(2)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(2),
			wantFingerprint: []string{ToHex(1)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(3),
			wantFingerprint: []string{ToHex(4)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(4),
			wantFingerprint: []string{ToHex(3)},
			wantNumSessions: 2,
			wantErr:         false,
		},
		{
			uid:             ToHex(5),
			wantFingerprint: []string{},
			wantNumSessions: 2,
			wantErr:         false,
		},
		// transaction does not exist in graph
		{
			uid:             ToHex(100),
			wantFingerprint: []string{},
			wantNumSessions: 2,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		fingerprints, i, err := SpendingFingerprint(g, tt.uid)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.wantNumSessions, i)
			fingerprintUIDs := make([]string, len(fingerprints))
			for j, f := range fingerprints {
				fingerprintUIDs[j] = f.TransactionUID
			}

			sort.Strings(fingerprintUIDs)

			require.Equal(t, tt.wantFingerprint, fingerprintUIDs)
		}
	}
}
