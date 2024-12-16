package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRevealHash(t *testing.T) {
	rb := RevealBody{
		ID:           "8686b3e5f77755a52652d78482c748a3a4e1e8f83befcfa1fc8e1f92197fe0b0",
		Salt:         []byte("salt"),
		Reveal:       "Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w=",
		GasUsed:      0,
		ExitCode:     0,
		ProxyPubKeys: []string{"key1", "key2"},
	}
	hash, err := rb.TryHash()
	require.NoError(t, err)
	require.Equal(t, "20a9224cd19a10ffb63c4b257e4f992f5d9042af142e03e60844adb9945f7d9b", hash)
}
