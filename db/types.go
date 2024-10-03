package db

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
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
	SearchTerm      string   `form:"q"`
	IscnId          string   `form:"iscn_id"`
	IscnIdPrefix    string   `form:"iscn_id_prefix"`
	Owner           string   `form:"owner"`
	Keywords        []string `form:"keyword"`
	Fingerprints    []string `form:"fingerprint"`
	StakeholderId   string   `form:"stakeholder.id"`
	StakeholderName string   `form:"stakeholder.name"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
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
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Symbol         string          `json:"symbol"`
	URI            string          `json:"uri"`
	URIHash        string          `json:"uri_hash"`
	Config         json.RawMessage `json:"config"`
	Metadata       json.RawMessage `json:"metadata"`
	Parent         NftClassParent  `json:"parent"`
	CreatedAt      time.Time       `json:"created_at"`
	LatestPrice    uint64          `json:"latest_price,omitempty"`
	PriceUpdatedAt *time.Time      `json:"price_updated_at,omitempty"`
}

type NftClassParent struct {
	Type         string `json:"type"`
	IscnIdPrefix string `json:"iscn_id_prefix"`
	Account      string `json:"account"`
}

type NoTimeZoneTime struct {
	time.Time
}

const NoTimeZoneTimeLayout = "2006-01-02T15:04:05"

func (ntzt *NoTimeZoneTime) UnmarshalJSON(b []byte) error {
	t, err := time.Parse(NoTimeZoneTimeLayout, strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}
	ntzt.Time = t
	return nil
}

type Nft struct {
	NftId          string          `json:"nft_id"`
	ClassId        string          `json:"class_id"`
	Owner          string          `json:"owner"`
	Uri            string          `json:"uri"`
	UriHash        string          `json:"uri_hash"`
	Metadata       json.RawMessage `json:"metadata"`
	Timestamp      time.Time       `json:"timestamp"`
	LatestPrice    uint64          `json:"latest_price,omitempty"`
	PriceUpdatedAt *NoTimeZoneTime `json:"price_updated_at,omitempty"`
}

type NftEventAction string

const (
	ACTION_SEND         NftEventAction = "/cosmos.nft.v1beta1.MsgSend"
	ACTION_MINT         NftEventAction = "mint_nft"
	ACTION_NEW_CLASS    NftEventAction = "new_class"
	ACTION_UPDATE_CLASS NftEventAction = "update_class"
	ACTION_BUY          NftEventAction = "buy_nft"
	ACTION_SELL         NftEventAction = "sell_nft"
)

type NftEvent struct {
	Action    NftEventAction     `json:"action"`
	ClassId   string             `json:"class_id"`
	NftId     string             `json:"nft_id"`
	Sender    string             `json:"sender"`
	Receiver  string             `json:"receiver"`
	MsgIndex  int                `json:"msg_index,omitempty"`
	Events    types.StringEvents `json:"events,omitempty"`
	TxHash    string             `json:"tx_hash"`
	Timestamp time.Time          `json:"timestamp"`
	Memo      string             `json:"memo"`
	Price     uint64             `json:"price,omitempty"`
}

type NftMarketplaceItem struct {
	Type       string    `json:"action,omitempty"`
	ClassId    string    `json:"class_id"`
	NftId      string    `json:"nft_id"`
	Creator    string    `json:"creator"`
	Price      uint64    `json:"price,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
}

type NftIncome struct {
	ClassId   string `json:"class_id"`
	NftId     string `json:"nft_id"`
	TxHash    string `json:"tx_hash"`
	Address   string `json:"address"`
	Amount    uint64 `json:"amount"`
	IsRoyalty bool   `json:"is_royalty"`
}

type LegacyPageRequest struct {
	Key     uint64 `form:"key"`
	Limit   int    `form:"limit,default=100" binding:"gte=1,lte=100"`
	Reverse bool   `form:"reverse"`
	Offset  uint64 `form:"offset"`
}

type PageRequest struct {
	Key     uint64 `form:"pagination.key"`
	Limit   int    `form:"pagination.limit,default=100" binding:"gte=1,lte=100"`
	Reverse bool   `form:"pagination.reverse"`
	Offset  uint64 `form:"pagination.offset"`
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
	Total   int    `json:"total,omitempty"`
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
	IscnIdPrefix    string   `form:"iscn_id_prefix"`
	Account         string   `form:"account"`
	IscnOwner       []string `form:"iscn_owner"`
	Owner           string   `form:"owner"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
}

type QueryClassResponse struct {
	Classes    []NftClassResponse `json:"classes"`
	Pagination PageResponse       `json:"pagination"`
}

type NftClassResponse struct {
	NftClass
	Owner          string     `json:"owner"`
	NftOwnedCount  *int       `json:"nft_owned_count,omitempty"`
	NftLastOwnedAt *time.Time `json:"nft_last_owned_at,omitempty"`
	LastOwnedNftId *string    `json:"last_owned_nft_id,omitempty"`
}

type QueryNftRequest struct {
	Owner         string `form:"owner" binding:"required"`
	ExpandClasses bool   `form:"expand_classes"`
}

type QueryNftResponse struct {
	Pagination PageResponse  `json:"pagination"`
	Nfts       []NftResponse `json:"nfts"`
}

type NftResponse struct {
	Nft
	ClassParent NftClassParent `json:"class_parent"`
	ClassData   *NftClass      `json:"class_data,omitempty"`
}

type QueryOwnerRequest struct {
	ClassId          string   `form:"class_id" binding:"required"`
	ExcludeIscnOwner bool     `form:"exclude_iscn_owner"`
	IgnoreList       []string `form:"ignore_list"`
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
	ClassId        string           `form:"class_id"`
	NftId          string           `form:"nft_id"`
	IscnIdPrefix   string           `form:"iscn_id_prefix"`
	Sender         []string         `form:"sender"`
	Receiver       []string         `form:"receiver"`
	Creator        []string         `form:"creator"`
	Involver       []string         `form:"involver"`
	Verbose        bool             `form:"verbose"`
	ActionType     []NftEventAction `form:"action_type"`
	IgnoreFromList []string         `form:"ignore_from_list"`
	IgnoreToList   []string         `form:"ignore_to_list"`
}

type QueryEventsResponse struct {
	Pagination PageResponse `json:"pagination"`
	Events     []NftEvent   `json:"events"`
}

type QueryIncomesRequest struct {
	ClassId             string           `form:"class_id"`
	Owner               string           `form:"owner"`
	Address             string           `form:"address"`
	After               int64            `form:"after"`
	Before              int64            `form:"before"`
	ActionType          []NftEventAction `form:"action_type"`
	IsIscnOwner         *bool            `form:"is_iscn_owner"`
	IsRoyalty           *bool            `form:"is_royalty"`
	ExcludeSelfPurchase bool             `form:"exclude_self_purchase"`
	OrderBy             string           `form:"order_by"`
}

type NftIncomeResponse struct {
	Address   string `json:"address"`
	Amount    uint64 `json:"amount"`
	IsRoyalty bool   `json:"is_royalty"`
}

type NftClassIncomeResponse struct {
	ClassId     string              `json:"class_id"`
	CreatedAt   time.Time           `json:"created_at"`
	Sales       uint64              `json:"sales"`
	TotalAmount uint64              `json:"total_amount"`
	Incomes     []NftIncomeResponse `json:"incomes"`
}

type QueryIncomesResponse struct {
	TotalSales   uint64                   `json:"total_sales"`
	TotalAmount  uint64                   `json:"total_amount"`
	ClassIncomes []NftClassIncomeResponse `json:"class_incomes"`
	Pagination   PageResponse             `json:"pagination"`
}

type QueryRankingRequest struct {
	StakeholderId   string   `form:"stakeholder_id"`
	StakeholderName string   `form:"stakeholder_name"`
	Creator         string   `form:"creator"`
	Type            string   `form:"type"`
	Collector       string   `form:"collector"`
	CreatedAfter    int64    `form:"created_after"`
	CreatedBefore   int64    `form:"created_before"`
	IncludeOwner    bool     `form:"include_owner"`
	IgnoreList      []string `form:"ignore_list"`
	ApiAddresses    []string `form:"api_addresses"`
	After           int64    `form:"after"`
	Before          int64    `form:"before"`
	OrderBy         string   `form:"order_by"`
}

type QueryRankingResponse struct {
	Classes    []NftClassRankingResponse `json:"classes"`
	Pagination PageResponse              `json:"pagination"`
}

type NftClassRankingResponse struct {
	NftClass
	Owner          string `json:"owner"`
	SoldCount      int    `json:"sold_count"`
	TotalSoldValue int64  `json:"total_sold_value"`
}

type QueryCollectorRequest struct {
	Creator         string   `form:"creator"`
	IgnoreList      []string `form:"ignore_list"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
	IncludeOwner    bool     `form:"include_owner,default=true"`
	PriceBy         string   `form:"price_by,default=nft"`
	OrderBy         string   `form:"order_by,default=price"`
}

type QueryCollectorResponse struct {
	Collectors []accountCollection `json:"collectors"`
	Pagination PageResponse        `json:"pagination"`
}

type QueryCreatorRequest struct {
	Collector       string   `form:"collector"`
	IgnoreList      []string `form:"ignore_list"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
	IncludeOwner    bool     `form:"include_owner,default=true"`
	PriceBy         string   `form:"price_by,default=nft"`
	OrderBy         string   `form:"order_by,default=price"`
}

type QueryCreatorResponse struct {
	Creators   []accountCollection `json:"creators"`
	Pagination PageResponse        `json:"pagination"`
}

type accountCollection struct {
	Account     string       `json:"account"`
	TotalValue  uint64       `json:"total_value"`
	Count       int          `json:"count"`
	Collections []collection `json:"collections"`
}

type collection struct {
	IscnIdPrefix string `json:"iscn_id_prefix"`
	ClassId      string `json:"class_id"`
	Value        int    `json:"value"`
	Count        int    `json:"count"`
}

type QueryUserStatRequest struct {
	User            string   `form:"user"`
	IgnoreList      []string `form:"ignore_list"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
}

type QueryUserStatResponse struct {
	CollectedClasses []CollectedClass `json:"collected_classes"`
	CreatedCount     int              `json:"created_count"`
	CollectorCount   int              `json:"collector_count"`
	TotalSales       uint64           `json:"total_sales"`
	TotalIncomes     uint64           `json:"total_incomes"`
}

type CollectedClass struct {
	ClassId string `json:"class_id"`
	Count   int    `json:"count"`
}

type QueryCountResponse struct {
	Count uint64 `json:"count"`
}

type QueryNftReturningCreatorCountRequest struct {
	ReturningThresholdDays int    `form:"returning_threshold_days"`
	Interval               string `form:"interval"`
	After                  int64  `form:"after"`
	Before                 int64  `form:"before"`
}

type NftReturningCreatorCountResponse struct {
	StartAt               time.Time `json:"start_at"`
	NewCreatorCount       uint64    `json:"new_creator_count"`
	ReturningCreatorCount uint64    `json:"returning_creator_count"`
}

type QueryNftReturningCreatorCountResponse struct {
	Intervals []NftReturningCreatorCountResponse `json:"intervals"`
}

type QueryNftCountRequest struct {
	IncludeOwner bool     `form:"include_owner"`
	IgnoreList   []string `form:"ignore_list"`
}

type QueryNftTradeStatsRequest struct {
}

type QueryNftTradeStatsResponse struct {
	Count       uint64 `json:"count"`
	TotalVolume uint64 `json:"total_volume"`
}

type QueryNftOwnerListResponse struct {
	Owners     []OwnerResponse `json:"owners"`
	Pagination PageResponse    `json:"pagination"`
}

type QueryNftMarketplaceItemsRequest struct {
	Type    string `form:"type"`
	ClassId string `form:"class_id"`
	NftId   string `form:"nft_id"`
	Creator string `form:"creator"`
	Expand  bool   `form:"expand"`
}

type NftMarketplaceItemResponse struct {
	NftMarketplaceItem
	ClassMetadata json.RawMessage `json:"class_metadata"`
	NftMetadata   json.RawMessage `json:"nft_metadata"`
}

type QueryNftMarketplaceItemsResponse struct {
	Items      []NftMarketplaceItemResponse `json:"items"`
	Pagination PageResponse                 `json:"pagination"`
}

type QueryCollectorTopRankedCreatorsRequest struct {
	Collector       string   `form:"collector" binding:"required"`
	IgnoreList      []string `form:"ignore_list"`
	AllIscnVersions bool     `form:"all_iscn_versions"`
	IncludeOwner    bool     `form:"include_owner,default=true"`
	Top             uint     `form:"top,default=5"`
}

type CollectorTopRankedCreator struct {
	Creator string `json:"creator"`
	Rank    uint   `json:"rank"`
}

type QueryCollectorTopRankedCreatorsResponse struct {
	Creators []CollectorTopRankedCreator `json:"creators"`
}

type QueryClassesOwnersRequest struct {
	ClassIds []string `form:"class_ids" binding:"required"`
	Owners   []string `form:"owners"`
}

type QueryClassesOwnersResponse struct {
	// key: owner address, value: class IDs
	Owners map[string][]string `json:"owners"`
}
