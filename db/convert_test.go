package db

import "testing"

func TestConvert(t *testing.T) {
	p := Pagination{
		Limit: 10,
		Page:  1,
		Order: ORDER_ASC,
	}

	err := ConvertISCN(pool, p)
	if err != nil {
		t.Error(err)
	}
}
