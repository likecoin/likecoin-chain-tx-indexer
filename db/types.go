package db

import (
	"encoding/json"
	"time"
)

type ISCN struct {
	Iscn         string
	IscnPrefix   string
	Owner        string
	Timestamp    string
	Ipld         string
	Keywords     []string
	Fingerprints []string
	Stakeholders []byte
	Data         []byte
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
