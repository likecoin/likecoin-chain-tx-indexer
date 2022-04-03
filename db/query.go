package db

import "encoding/json"

type ISCNRecordQuery struct {
	ContentFingerprints []string         `json:"contentFingerprints,omitempty"`
	ContentMetadata     *ContentMetadata `json:"contentMetadata,omitempty"`
	Stakeholders        []Stakeholder    `json:"stakeholders,omitempty"`
}

type ContentMetadata struct {
	Name     string `json:"name,omitempty"`
	Keywords string `json:"keywords,omitempty"`
	Type     string `json:"@type,omitempty"`
}

type Stakeholder struct {
	Entity *Entity `json:"entity,omitempty"`
}

type Entity struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (q ISCNRecordQuery) Marshal() ([]byte, error) {
	return json.Marshal(q)
}

type Pagination struct {
	Limit uint
	Page  uint
	Order Order
}

func (p Pagination) getOffset() uint {
	return p.Limit * (p.Page)
}
