package constants

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrivacyType_IsCC(t *testing.T) {
	tests := []struct {
		t    PrivacyType
		want bool
	}{
		{t: PrivacyType(0), want: false},
		{t: PrivacyType(301), want: true},
		{t: PrivacyType(100), want: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.t.IsCC())
	}
}

func TestPrivacyType_IsCP(t *testing.T) {
	tests := []struct {
		t    PrivacyType
		want bool
	}{
		{t: PrivacyType(0), want: false},
		{t: PrivacyType(401), want: true},
		{t: PrivacyType(100), want: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.t.IsCP())
	}
}

func TestPrivacyType_IsDestination(t *testing.T) {
	tests := []struct {
		t    PrivacyType
		want bool
	}{
		{t: PrivacyType(0), want: false},
		{t: PrivacyType(101), want: true},
		{t: PrivacyType(100), want: true},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.t.IsDestination())
	}
}

func TestPrivacyType_IsMixing(t *testing.T) {
	tests := []struct {
		t    PrivacyType
		want bool
	}{
		{t: PrivacyType(0), want: true},
		{t: PrivacyType(301), want: false},
		{t: PrivacyType(100), want: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.t.IsMixing())
	}
}

func TestPrivacyType_IsOrigin(t *testing.T) {
	tests := []struct {
		t    PrivacyType
		want bool
	}{
		{t: PrivacyType(0), want: false},
		{t: PrivacyType(201), want: true},
		{t: PrivacyType(100), want: false},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.t.IsOrigin())
	}
}
