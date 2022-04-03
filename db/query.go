package db

import "encoding/json"

type iscnRecordQuery struct {
	ContentFingerprints []string         `json:"contentFingerprints,omitempty"`
	ContentMetadata     *contentMetadata `json:"contentMetadata,omitempty"`
}

type contentMetadata struct {
	Name     string `json:"name,omitempty"`
	Keywords string `json:"keywords,omitempty"`
	Type     string `json:"@type,omitempty"`
}

func marshallQuery(query iscnRecordQuery) ([]byte, error) {
	return json.Marshal(query)
}
