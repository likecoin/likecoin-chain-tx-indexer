package db

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type Stakeholder struct {
	Entity Entity `json:"entity,omitempty"`
	Data   []byte
}

type Entity struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (e *Entity) UnmarshalJSON(data []byte) (err error) {
	dict := make(map[string]interface{})
	if err = json.Unmarshal(data, &dict); err != nil {
		return
	}
	if v, ok := dict["id"].(string); ok {
		e.Id = v
	}
	if v, ok := dict["@id"].(string); ok {
		e.Id = v
	}
	if v, ok := dict["name"].(string); ok {
		e.Name = v
	}
	return nil
}

type IscnInsert struct {
	Iscn         string
	IscnPrefix   string
	Version      int
	Owner        string
	Timestamp    time.Time
	Ipld         string
	Name         string
	Description  string
	Url          string
	Keywords     []string
	Fingerprints []string
	Stakeholders []Stakeholder
	Data         []byte
}

type IscnQuery struct {
	IscnId          string   `form:"iscn_id"`
	IscnIdPrefix    string   `form:"iscn_id_prefix"`
	Owner           string   `form:"owner"`
	Keywords        []string `form:"keyword"`
	Fingerprints    []string `form:"fingerprint"`
	StakeholderId   string   `form:"stakeholder.id"`
	StakeholderName string   `form:"stakeholder.name"`
}

func (q IscnQuery) Empty() bool {
	return q.IscnId == "" &&
		q.IscnIdPrefix == "" &&
		q.Owner == "" &&
		len(q.Keywords) == 0 &&
		len(q.Fingerprints) == 0 &&
		q.StakeholderId == "" &&
		q.StakeholderName == ""
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
	CreatedAt   time.Time       `json:"created_at"`
}

type NftClassParent struct {
	Type         string `json:"type"`
	IscnIdPrefix string `json:"iscn_id_prefix"`
	Account      string `json:"account"`
}

type Nft struct {
	NftId     string          `json:"nft_id"`
	ClassId   string          `json:"class_id"`
	Owner     string          `json:"owner"`
	Uri       string          `json:"uri"`
	UriHash   string          `json:"uri_hash"`
	Metadata  json.RawMessage `json:"metadata"`
	Timestamp time.Time       `json:"timestamp"`
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

type PageRequest struct {
	Key     uint64 `form:"key"`
	Limit   int    `form:"limit,default=100" binding:"gte=1,lte=100"`
	Reverse bool   `form:"reverse"`
	Offset  uint64 `form:"offset"`
}

func (p *PageRequest) After() uint64 {
	if p.Reverse {
		return 0
	}
	return p.Key
}

func (p *PageRequest) Before() uint64 {
	if p.Reverse {
		return p.Key
	}
	return 0
}

func (p *PageRequest) Order() Order {
	if p.Reverse {
		return ORDER_DESC
	}
	return ORDER_ASC
}

type PageResponse struct {
	NextKey uint64 `json:"next_key,omitempty"`
	Count   int    `json:"count,omitempty"`
}

type IscnResponse struct {
	Records    []iscnResponseRecord `json:"records"`
	Pagination PageResponse         `json:"pagination"`
}

type iscnResponseRecord struct {
	Ipld string           `json:"ipld,omitempty"`
	Data iscnResponseData `json:"data,omitempty"`
}

type iscnResponseData struct {
	Id                  string          `json:"@id"`
	RecordTimestamp     time.Time       `json:"recordTimestamp"`
	Owner               string          `json:"owner"`
	RecordNotes         json.RawMessage `json:"recordNotes"`
	ContentFingerprints json.RawMessage `json:"contentFingerprints,omitempty"`
	ContentMetadata     json.RawMessage `json:"contentMetadata,omitempty"`
	Stakeholders        json.RawMessage `json:"stakeholders,omitempty"`
}

type QueryClassRequest struct {
	IscnIdPrefix string `form:"iscn_id_prefix"`
	Account      string `form:"account"`
	Expand       bool   `form:"expand"`
}

type QueryClassResponse struct {
	Classes    []NftClassResponse `json:"classes"`
	Pagination PageResponse       `json:"pagination"`
}

type NftClassResponse struct {
	NftClass
	Count int   `json:"count,omitempty"`
	Nfts  []Nft `json:"nfts,omitempty"`
}

type QueryNftRequest struct {
	Owner string `form:"owner" binding:"required"`
}

type QueryNftResponse struct {
	Pagination PageResponse  `json:"pagination"`
	Nfts       []NftResponse `json:"nfts"`
}

type NftResponse struct {
	Nft
	ClassParent NftClassParent `json:"class_parent"`
}

type QueryOwnerRequest struct {
	ClassId string `form:"class_id" binding:"required"`
}

type QueryOwnerResponse struct {
	Pagination PageResponse    `json:"pagination"`
	Owners     []OwnerResponse `json:"owners"`
}

type OwnerResponse struct {
	Owner string   `json:"owner"`
	Count int      `json:"count,omitempty"`
	Nfts  []string `json:"nfts,omitempty"`
}

type QueryEventsRequest struct {
	ClassId      string   `form:"class_id"`
	NftId        string   `form:"nft_id"`
	IscnIdPrefix string   `form:"iscn_id_prefix"`
	Verbose      bool     `form:"verbose"`
	ActionType   []string `form:"action_type"`
}

type QueryEventsResponse struct {
	Pagination PageResponse `json:"pagination"`
	Events     []NftEvent   `json:"events"`
}

type QueryRankingRequest struct {
	StakeholderId   string   `form:"stakeholder_id"`
	StakeholderName string   `form:"stakeholder_name"`
	Creator         string   `form:"creator"`
	Type            string   `form:"type"`
	Collector       string   `form:"collector"`
	After           int64    `form:"after"`
	Before          int64    `form:"before"`
	IncludeOwner    bool     `form:"include_owner"`
	IgnoreList      []string `form:"ignore_list"`
	PageRequest
}

type QueryRankingResponse struct {
	Classes    []NftClassRankingResponse `json:"classes"`
	Pagination PageResponse              `json:"pagination"`
}

type NftClassRankingResponse struct {
	NftClass
	SoldCount int `json:"sold_count"`
}

type QueryCollectorRequest struct {
	Creator    string   `form:"creator"`
	IgnoreList []string `form:"ignore_list"`
	PageRequest
}

type QueryCollectorResponse struct {
	Collectors []accountCollection `json:"collectors"`
	Pagination PageResponse        `json:"pagination"`
}

type QueryCreatorRequest struct {
	Collector  string   `form:"collector"`
	IgnoreList []string `form:"ignore_list"`
	PageRequest
}

type QueryCreatorResponse struct {
	Creators   []accountCollection `json:"creators"`
	Pagination PageResponse        `json:"pagination"`
}

type accountCollection struct {
	Account     string       `json:"account"`
	Count       int          `json:"count"`
	Collections []collection `json:"collections"`
}

type collection struct {
	IscnIdPrefix string `json:"iscn_id_prefix"`
	ClassId      string `json:"class_id"`
	Count        int    `json:"count"`
}

type QueryCountResponse struct {
	Count uint64 `json:"count"`
}

type QueryNftCountRequest struct {
	IncludeOwner bool     `form:"include_owner"`
	IgnoreList   []string `form:"ignore_list"`
}

type QueryNftTradeStatsRequest struct {
	ApiAddress string `form:"api_address" binding:"required"`
}

type QueryNftTradeStatsResponse struct {
	Count       uint64 `json:"count"`
	TotalVolume uint64 `json:"total_volume"`
}

type QueryNftOwnerListResponse struct {
	Owners     []OwnerResponse `json:"owners"`
	Pagination PageResponse    `json:"pagination"`
}
