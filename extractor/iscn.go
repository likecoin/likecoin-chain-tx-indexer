package extractor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/likecoin/likecoin-chain-tx-indexer/db"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type iscnMessage struct {
	Record json.RawMessage `json:"record"`
}

type iscnData struct {
	ContentFingerprints []string         `json:"contentFingerprints,omitempty"`
	ContentMetadata     *contentMetadata `json:"contentMetadata,omitempty"`
	Stakeholders        []stakeholder    `json:"stakeholders,omitempty"`
}

type contentMetadata struct {
	Keywords string `json:"keywords,omitempty"`
}

type stakeholder struct {
	Entity *entity `json:"entity,omitempty"`
}

type entity struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (e *entity) UnmarshalJSON(data []byte) (err error) {
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

func (q iscnData) Marshal() ([]byte, error) {
	return json.Marshal(q)
}

func insertISCN(payload db.EventPayload) error {
	message := payload.Message
	events := payload.Events
	var data iscnMessage
	if err := json.Unmarshal(message, &data); err != nil {
		return fmt.Errorf("Failed to unmarshal iscn: %w", err)
	}
	var record iscnData
	if err := json.Unmarshal(data.Record, &record); err != nil {
		return fmt.Errorf("Failed to unmarshal iscn: %w", err)
	}
	holders, err := formatStakeholders(record.Stakeholders)
	if err != nil {
		return fmt.Errorf("Failed to format stakeholder, %w", err)
	}
	iscn := db.ISCN{
		Iscn:         utils.GetEventsValue(events, "iscn_record", "iscn_id"),
		IscnPrefix:   utils.GetEventsValue(events, "iscn_record", "iscn_id_prefix"),
		Version:      getIscnVersion(utils.GetEventsValue(events, "iscn_record", "iscn_id")),
		Owner:        utils.GetEventsValue(events, "iscn_record", "owner"),
		Keywords:     utils.ParseKeywords(record.ContentMetadata.Keywords),
		Fingerprints: record.ContentFingerprints,
		Stakeholders: holders,
		Timestamp:    payload.Timestamp,
		Ipld:         utils.GetEventsValue(events, "iscn_record", "ipld"),
		Data:         data.Record,
	}
	payload.Batch.InsertISCN(iscn)
	return nil
}

func transferISCN(payload db.EventPayload) error {
	events := payload.Events
	sender := utils.GetEventsValue(events, "message", "sender")
	iscnId := utils.GetEventsValue(events, "iscn_record", "iscn_id")
	newOwner := utils.GetEventsValue(events, "iscn_record", "owner")
	payload.Batch.Batch.Queue(`UPDATE iscn SET owner = $2 WHERE iscn_id = $1`, iscnId, newOwner)
	logger.L.Debugf("Send ISCN %s from %s to %s\n", iscnId, sender, newOwner)
	return nil
}

func getIscnVersion(iscn string) int {
	arr := strings.Split(iscn, "/")
	last := arr[len(arr)-1]
	result, err := strconv.Atoi(last)
	if err != nil {
		return 0
	}
	return result
}

func formatStakeholders(stakeholders []stakeholder) ([]byte, error) {
	body := make([]*entity, len(stakeholders))
	for i, v := range stakeholders {
		body[i] = v.Entity
	}
	return json.Marshal(body)
}
