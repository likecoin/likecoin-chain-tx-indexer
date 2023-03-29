package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertAddressPrefixes(t *testing.T) {
	prefixes := []string{"like", "cosmos"}
	addr1 := "like1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wz9ykdd5"
	addr2 := "cosmos1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wzkc20w0"
	expected := []string{addr1, addr2}

	convertedAddrs1 := ConvertAddressPrefixes(addr1, prefixes)
	require.Equal(t, convertedAddrs1, expected)

	convertedAddrs2 := ConvertAddressPrefixes(addr2, prefixes)
	require.Equal(t, convertedAddrs2, expected)

	wrongAddr := "like1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wz9ykdd6" // wrong checksum
	convertedAddrs3 := ConvertAddressPrefixes(wrongAddr, prefixes)
	expected = []string{wrongAddr}
	require.Equal(t, convertedAddrs3, expected)

	convertedAddrs4 := ConvertAddressPrefixes("", prefixes)
	require.Empty(t, convertedAddrs4)
}
