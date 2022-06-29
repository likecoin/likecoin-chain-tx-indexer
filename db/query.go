package db

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ISCNRecordQuery struct {
	ContentFingerprints []string         `json:"contentFingerprints,omitempty"`
	ContentMetadata     *ContentMetadata `json:"contentMetadata,omitempty"`
	Stakeholders        []Stakeholder    `json:"stakeholders,omitempty"`
}

type ContentMetadata struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"@type,omitempty"`
	Keywords string `json:"keywords,omitempty"`
}

type Stakeholder struct {
	Entity *Entity `json:"entity,omitempty"`
}

type Entity struct {
	Id   string `json:"@id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (q ISCNRecordQuery) Marshal() ([]byte, error) {
	return json.Marshal(q)
}

type Pagination struct {
	Limit uint64
	Page  uint64
	Order Order
}

func (p Pagination) getOffset() uint64 {
	return p.Limit * (p.Page - 1)
}

type Keywords []string

func (k Keywords) Marshal() string {
	if k == nil {
		return "{}"
	}
	body, err := json.Marshal(k)
	if err != nil {
		return "{}"
	}
	s := strings.Trim(string(body), "[]")
	return fmt.Sprintf("{%s}", s)
}
