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
	Keywords string `json:"keywords,omitempty"`
}

type Stakeholder struct {
	Entity *Entity `json:"entity,omitempty"`
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
