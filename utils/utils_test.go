package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseKeywords(t *testing.T) {
	table := map[string][]string{
		"":      {},
		", ":    {},
		"a, b":  {"a", "b"},
		"a,bc":  {"a", "bc"},
		"a,bc,": {"a", "bc"},
		"  ":    {},
	}
	for k, v := range table {
		require.Equal(t, ParseKeywords(k), v)
	}
}
