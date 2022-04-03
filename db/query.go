package db

import "encoding/json"

type iscnRecordQuery struct {
	ContentFingerprints []string         `json:"contentFingerprints,omitempty"`
	ContentMetadata     *contentMetadata `json:"contentMetadata,omitempty"`
	Stakeholders        []stakeholder    `json:"stakeholders,omitempty"`
}

type contentMetadata struct {
	Name     string `json:"name,omitempty"`
	Keywords string `json:"keywords,omitempty"`
	Type     string `json:"@type,omitempty"`
}

type stakeholder struct {
	Entity *entity `json:"entity,omitempty"`
}

type entity struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func marshallQuery(query iscnRecordQuery) ([]byte, error) {
	return json.Marshal(query)
}
