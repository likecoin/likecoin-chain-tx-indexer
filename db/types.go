package db

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
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
	Action    string             `json:"action"`
	ClassId   string             `json:"class_id"`
	NftId     string             `json:"nft_id"`
	Sender    string             `json:"sender"`
	Receiver  string             `json:"receiver"`
	Events    types.StringEvents `json:"events,omitempty"`
	TxHash    string             `json:"tx_hash"`
	Timestamp time.Time          `json:"timestamp"`
}

func (n *NftEvent) Attach(payload EventPayload) {
	n.Action = utils.GetEventsValue(payload.Events, "message", "action")
	n.Events = payload.Events
	n.Timestamp = payload.Timestamp
	n.TxHash = payload.TxHash
}

type Pagination struct {
	After  uint64
	Before uint64
	Limit  uint64
	Order  Order
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

type QueryClassRequest struct {
	IscnIdPrefix string `form:"iscn_id_prefix" binding:"required"`
	Expand       bool   `form:"expand"`
}

type QueryClassResponse struct {
	Classes []NftClassResponse `json:"classes"`
}

type NftClassResponse struct {
	NftClass
	Count int   `json:"count"`
	Nfts  []Nft `json:"nfts,omitempty"`
}

type QueryNftRequest struct {
	Owner string `form:"owner" binding="required"`
}

type QueryNftResponse struct {
	Nfts []NftResponse `json:"nfts"`
}

type NftResponse struct {
	Nft
	ClassParent NftClassParent `json:"class_parent"`
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

type QueryEventsRequest struct {
	ClassId      string `form:"class_id"`
	NftId        string `form:"nft_id"`
	IscnIdPrefix string `form:"iscn_id_prefix"`
	Verbose      bool   `form:"verbose"`
}

type QueryEventsResponse struct {
	Count  int        `json:"count"`
	Events []NftEvent `json:"events"`
}
