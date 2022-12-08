package utils

import "testing"

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

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
		if ans := ParseKeywords(k); !equal(ans, v) {
			t.Errorf("parse %#v expect %#v got %#v", k, v, ans)
		}
	}
}
