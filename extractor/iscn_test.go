package extractor

import (
	"testing"
)

func TestIscnVersion(t *testing.T) {
	table := []struct {
		iscn    string
		version int
	}{
		{
			iscn:    "iscn://likecoin-chain/Nj8mKU_TnRFp5kytMF7hJk4_unujhqM0V_9gFrleAgs/1",
			version: 1,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U/3",
			version: 3,
		},
		{
			iscn:    "iscn://likecoin-chain/vxhbRBaMGSdpgaYp7gk7y8iTDMlc6QVZ6XzxaLKGa0U",
			version: 0,
		},
	}
	for _, v := range table {
		if a := getIscnVersion(v.iscn); a != v.version {
			t.Errorf("parse %s expect %d got %d\n", v.iscn, v.version, a)
		}
	}
}
