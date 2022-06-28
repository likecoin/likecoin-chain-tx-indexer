package db

import "testing"

func TestConvert(t *testing.T) {
	err := ConvertISCN(pool, 1000, 20)
	if err != nil {
		t.Error(err)
	}
}
