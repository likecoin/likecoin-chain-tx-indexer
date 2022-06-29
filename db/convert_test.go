package db

import "testing"

func TestConvert(t *testing.T) {
	conn, err := AcquireFromPool(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Release()
	for i := 0; i < 2; i++ {
		err := ConvertISCN(conn, 50000)
		if err != nil {
			t.Error(err)
		}
	}
}
