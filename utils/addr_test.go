package utils

import "testing"

func TestConvertAddressPrefixes(t *testing.T) {
	prefixes := []string{"like", "cosmos"}
	addr1 := "like1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wz9ykdd5"
	addr2 := "cosmos1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wzkc20w0"
	expected := []string{addr1, addr2}

	convertedAddrs1 := ConvertAddressPrefixes(addr1, prefixes)
	if len(convertedAddrs1) != len(expected) {
		t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs1)
	}
	for i := range convertedAddrs1 {
		if convertedAddrs1[i] != expected[i] {
			t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs1)
		}
	}

	convertedAddrs2 := ConvertAddressPrefixes(addr2, prefixes)
	if len(convertedAddrs2) != len(expected) {
		t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs2)
	}
	for i := range convertedAddrs1 {
		if convertedAddrs2[i] != expected[i] {
			t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs2)
		}
	}

	wrongAddr := "like1hggde2u9lrjy9x9kqfwzzgjwkxe2y9wz9ykdd6" // wrong checksum
	convertedAddrs3 := ConvertAddressPrefixes(wrongAddr, prefixes)
	expected = []string{wrongAddr}
	if len(convertedAddrs3) != len(expected) {
		t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs3)
	}
	for i := range convertedAddrs3 {
		if convertedAddrs3[i] != expected[i] {
			t.Errorf("ConvertAddressPrefixes failed, expect %#v, got %#v", expected, convertedAddrs3)
		}
	}

	convertedAddrs4 := ConvertAddressPrefixes("", prefixes)
	if len(convertedAddrs4) != 0 {
		t.Errorf("ConvertAddressPrefixes failed, expect empty array, got %#v", convertedAddrs4)
	}
}
