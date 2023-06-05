package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSplitToParts(t *testing.T) {
	t.Parallel()

	type args struct {
		totalSize uint64
		stats     []float64
	}

	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "case1",
			args: args{totalSize: 100, stats: []float64{0.5, 0.5}},
			want: []uint64{50, 50},
		},
		{
			name: "case2",
			args: args{totalSize: 1, stats: []float64{0.5, 0.5}},
			want: []uint64{0, 1},
		},
		{
			name: "case3",
			args: args{totalSize: 1, stats: []float64{0.5, 0.6}},
			want: []uint64{0, 1},
		},
		{
			name: "case4",
			args: args{totalSize: 5, stats: []float64{0.773, 0.366, 0.662, 0.722, 0.842, 0.858}},
			want: []uint64{1, 0, 1, 1, 1, 1},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			split := splitToParts(tt.args.totalSize, tt.args.stats)
			require.Equal(t, len(tt.args.stats), len(split))
			require.Equal(t, tt.want, split)
		})
	}
}
