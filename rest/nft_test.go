package rest_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/rest"
	. "github.com/likecoin/likecoin-chain-tx-indexer/test"
)

func TestQueryClassesOwners(t *testing.T) {
	defer CleanupTestData(Conn)
	nftClasses := []NftClass{
		{
			Id: "nftlike1aaaaa1",
		},
		{
			Id: "nftlike1bbbbb1",
		},
		{
			Id: "nftlike1ccccc1",
		},
		{
			Id: "nftlike1ddddd1",
		},
	}
	nfts := []Nft{
		{
			NftId:   "testing-nft-109283748",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283749",
			ClassId: nftClasses[0].Id,
			Owner:   ADDR_02_COSMOS,
		},
		{
			NftId:   "testing-nft-109283750",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_03_LIKE,
		},
		{
			NftId:   "testing-nft-109283751",
			ClassId: nftClasses[1].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283752",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_01_LIKE,
		},
		{
			NftId:   "testing-nft-109283753",
			ClassId: nftClasses[2].Id,
			Owner:   ADDR_02_COSMOS,
		},
		{
			NftId:   "testing-nft-109283754",
			ClassId: nftClasses[3].Id,
			Owner:   "nobody",
		},
	}
	InsertTestData(DBTestData{NftClasses: nftClasses, Nfts: nfts})

	classIds012 := []string{nftClasses[0].Id, nftClasses[1].Id, nftClasses[2].Id}
	classIds02 := []string{nftClasses[0].Id, nftClasses[2].Id}
	classIds0 := []string{nftClasses[0].Id}
	classIds1 := []string{nftClasses[1].Id}

	testCases := []struct {
		name  string
		query QueryClassesOwnersRequest
		res   QueryClassesOwnersResponse
	}{
		{
			"class IDs (0, 1, 2) without owners",
			QueryClassesOwnersRequest{
				ClassIds: classIds012,
			},
			QueryClassesOwnersResponse{
				Owners: map[string][]string{
					ADDR_01_LIKE: classIds012,
					ADDR_02_LIKE: classIds02,
					ADDR_03_LIKE: classIds1,
				},
			},
		},
		{
			"class IDs (0, 1, 2), owners (01, 02)",
			QueryClassesOwnersRequest{
				ClassIds: classIds012,
				Owners:   []string{ADDR_01_LIKE, ADDR_02_LIKE},
			},
			QueryClassesOwnersResponse{
				Owners: map[string][]string{
					ADDR_01_LIKE: classIds012,
					ADDR_02_LIKE: classIds02,
				},
			},
		},
		{
			"class IDs (0), without owners",
			QueryClassesOwnersRequest{
				ClassIds: classIds0,
			},
			QueryClassesOwnersResponse{
				Owners: map[string][]string{
					ADDR_01_LIKE: classIds0,
					ADDR_02_LIKE: classIds0,
				},
			},
		},
		{
			"class IDs (0), owners (01, 03)",
			QueryClassesOwnersRequest{
				ClassIds: classIds0,
				Owners:   []string{ADDR_01_LIKE, ADDR_03_LIKE},
			},
			QueryClassesOwnersResponse{
				Owners: map[string][]string{
					ADDR_01_LIKE: classIds0,
				},
			},
		},
		{
			"class IDs (3), owners (01, 02, 03)",
			QueryClassesOwnersRequest{
				ClassIds: []string{nftClasses[3].Id},
				Owners:   []string{ADDR_01_LIKE, ADDR_02_LIKE, ADDR_03_LIKE},
			},
			QueryClassesOwnersResponse{
				Owners: map[string][]string{},
			},
		},
	}

	for i, testCase := range testCases {
		query := rest.NFT_ENDPOINT + "/classes-owners?"
		for _, classId := range testCase.query.ClassIds {
			query += "class_ids=" + classId + "&"
		}
		for _, owner := range testCase.query.Owners {
			query += "owners=" + owner + "&"
		}
		req := httptest.NewRequest("GET", query, nil)
		httpRes, body := request(req)
		require.Equal(t, 200, httpRes.StatusCode, "Error in test case #%02d (%s), query = `%s`, response = `%s`", i, testCase.name, testCase.query, body)
		var res QueryClassesOwnersResponse
		err := json.Unmarshal([]byte(body), &res)
		require.NoError(t, err, "Error in test case #%02d (%s), query = `%s` response = `%s`", i, testCase.name, testCase.query, body)
		require.Equal(t, testCase.res, res, "Error in test case #%02d (%s), query = `%s`, response = `%s`", i, testCase.name, testCase.query, body)
	}

	req := httptest.NewRequest(
		"GET",
		rest.NFT_ENDPOINT+"/classes-owners?owners="+ADDR_01_LIKE,
		nil,
	)
	httpRes, body := request(req)
	require.Equal(t, 400, httpRes.StatusCode)
	var res struct {
		Err string `json:"error"`
	}
	err := json.Unmarshal([]byte(body), &res)
	require.NoError(t, err)
	require.Contains(t, res.Err, "Field validation for 'ClassIds' failed on the 'required' tag")
}
