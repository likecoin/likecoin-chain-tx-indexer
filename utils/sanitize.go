package utils

import (
	"bytes"
	"strconv"
)

const (
	SANITIZER_OK = iota
	SANITIZER_BAD
)

func parse4Hex(bz []byte, start int) int {
	if start+4 > len(bz) {
		return -1
	}
	for n := 0; n < 4; n++ {
		curr := start + n
		b := bz[curr]
		// 0x30~0x39: '0'~'9'; 0x41~0x46: 'A'~'F'; 0x61~0x66: 'a'~'f'
		if !((b >= 0x30 && b <= 0x39) || (b >= 0x41 && b <= 0x46) || (b >= 0x61 && b <= 0x66)) {
			return -1
		}
	}
	n, err := strconv.ParseUint(string(bz[start:start+4]), 16, 16)
	if err != nil {
		// WTF?
		return -1
	}
	return int(n)
}

func parseEscape(bz []byte, start int) (sanitizerResult int, nextStart int) {
	// skip the '\'
	curr := start + 1
	if curr >= len(bz) {
		// Should not happen... But all problems should not happen in an ideal life
		return SANITIZER_BAD, curr
	}
	switch bz[curr] {
	case 0x22, 0x5C, 0x2F, 0x62, 0x66, 0x6E, 0x72, 0x74: // '"', '\', '/', 'b', 'f', 'n', 'r', 't'
		return SANITIZER_OK, curr + 1
	case 0x75: // 'u'. Uppercase version ("\Uxxxx") won't work.
		// If it succeeds, it must be 4 bytes
		curr++
		n1 := parse4Hex(bz, curr)
		if n1 < 0 {
			// Just return `curr+1` means we only consume the `\u`.
			return SANITIZER_BAD, curr
		}
		// the `\uxxxx` sequence may now be considered a whole
		curr += 4
		if n1 == 0 {
			// \u0000 is also not allowed, this time we remove the whole sequence
			return SANITIZER_BAD, curr
		}
		if n1 >= 0xDC00 && n1 <= 0xDFFF {
			// lower surrogate, should appear after a higher surrogate
			return SANITIZER_BAD, curr
		}
		if n1 >= 0xD800 && n1 <= 0xDBFF {
			// higher surrogate, should be followed by a lower surrogate
			if curr+6 > len(bz) || bz[curr] != 0x5C || bz[curr+1] != 0x75 {
				return SANITIZER_BAD, curr
			}
			curr += 2
			n2 := parse4Hex(bz, curr)
			if n2 < 0 {
				return SANITIZER_BAD, curr
			}
			curr += 4
			if n2 >= 0xDC00 && n2 <= 0xDFFF {
				return SANITIZER_OK, curr
			}
			return SANITIZER_BAD, curr
		}
		// other cases for \uXXXX, allowed
		return SANITIZER_OK, curr
	}
	// other characters after '\', not allowed in Postgres
	return SANITIZER_BAD, curr
}

// Need to remove \u0000 and Unicode surrogate pairs.
// \u0000 is not allowed in Postgres.
// JSON may contain UTF-16 sequences that is more than 16 bits, e.g. "\uD83D\uDE02" = "😂".
// These pairs are called surrogate pairs.
// They are converted into single character in Postgres JSONB storage.
// So Postgres don't accept one single dangling surrogate code point (they are actually invalid).
// But somehow these invalid code points appear in JSON...
func SanitizeJSON(bz []byte) []byte {
	// TODO: check the encoding of string used by the output of MarshalJSON.
	// If it is not UTF-8 (e.g. UTF-16), we cannot use []byte and need to process
	// strings and characters.
	output := &bytes.Buffer{}
	curr := 0
	for curr < len(bz) {
		switch bz[curr] {
		case 0x5C: // '\'
			result, nextStart := parseEscape(bz, curr)
			switch result {
			case SANITIZER_OK:
				// We may accumulate the consumed characters.
				curr = nextStart
			case SANITIZER_BAD:
				// We need to remove some bytes. This is done by:
				// 1. write the characters that previously passed
				// 2. reset bz so it skips after the bad characters consumed by parser
				// 3. since bz is reset, curr starts from 0
				output.Write(bz[0:curr])
				curr = 0
				bz = bz[nextStart:]
			}
		default:
			curr++
		}
	}
	output.Write(bz[0:curr])
	return output.Bytes()
}
