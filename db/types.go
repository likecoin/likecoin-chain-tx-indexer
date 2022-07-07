package db

import (
	"encoding/json"
	"time"
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
	NftId    string
	ClassId  string
	Owner    string
	Uri      string
	UriHash  string
	Metadata json.RawMessage
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
	Count int               `json:"count"`
	Nfts  []json.RawMessage `json:"nfts"`
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
