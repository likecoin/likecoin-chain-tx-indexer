package db

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
)

type ISCN struct {
	Iscn         string
	IscnPrefix   string
	Version      int
	Owner        string
	Timestamp    time.Time
	Ipld         string
	Keywords     []string
	Fingerprints []string
	Stakeholders []byte
	Data         []byte
}

type NftClass struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Symbol      string          `json:"symbol"`
	URI         string          `json:"uri"`
	URIHash     string          `json:"uri_hash"`
	Config      json.RawMessage `json:"config"`
	Metadata    json.RawMessage `json:"metadata"`
	Parent      NftClassParent  `json:"parent"`
	Price       int             `json:"price"`
}

type NftClassParent struct {
	Type         string `json:"type"`
	IscnIdPrefix string `json:"iscn_id_prefix"`
	Account      string `json:"account"`
}

type Nft struct {
	NftId    string          `json:"nft_id"`
	ClassId  string          `json:"class_id"`
	Owner    string          `json:"owner"`
	Uri      string          `json:"uri"`
	UriHash  string          `json:"uri_hash"`
	Metadata json.RawMessage `json:"metadata"`
}

type NftEvent struct {
	Action    string
	ClassId   string
	NftId     string
	Sender    string
	Receiver  string
	Events    types.StringEvents
	TxHash    string
	Timestamp time.Time
}

type Pagination struct {
	After  uint64
	Before uint64
	Limit  uint64
	Order  Order
}

type NftClassResponse struct {
	Classes []NftClass `json:"classes"`
}

type QueryNftByIscnResponse struct {
	IscnIdPrefix string                  `json:"iscn_id_prefix"`
	Classes      []QueryNftClassResponse `json:"classes"`
}

type QueryNftClassResponse struct {
	NftClass
	Count int   `json:"count"`
	Nfts  []Nft `json:"nfts"`
}

type QueryNftResponse struct {
	Nft
	ClassParent NftClassParent `json:"class_parent"`
}

type QueryNftByOwnerResponse struct {
	Owner string             `json:"owner"`
	Nfts  []QueryNftResponse `json:"nfts"`
}

type QueryOwnerByClassIdResponse struct {
	ClassId string               `json:"class_id"`
	Owners  []QueryOwnerResponse `json:"owners"`
}

type QueryOwnerResponse struct {
	Owner string   `json:"owner"`
	Count int      `json:"count"`
	Nfts  []string `json:"nfts"`
}

type ISCNResponse struct {
	Records []ISCNResponseRecord `json:"records"`
	Last    uint64               `json:"last"`
}

type ISCNResponseRecord struct {
	Ipld string           `json:"ipld,omitempty"`
	Data ISCNResponseData `json:"data,omitempty"`
}

type ISCNResponseData struct {
	Id                  string          `json:"@id"`
	RecordTimestamp     time.Time       `json:"recordTimestamp"`
	Owner               string          `json:"owner"`
	RecordNotes         json.RawMessage `json:"recordNotes"`
	ContentFingerprints json.RawMessage `json:"contentFingerprints,omitempty"`
	ContentMetadata     json.RawMessage `json:"contentMetadata,omitempty"`
	Stakeholders        json.RawMessage `json:"stakeholders,omitempty"`
}
