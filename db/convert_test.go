package db

import "testing"

func TestConvert(t *testing.T) {
	err := ConvertISCN(pool, 50000)
	if err != nil {
		t.Error(err)
	}
}
