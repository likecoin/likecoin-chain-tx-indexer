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
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Keywords    string `json:"keywords,omitempty"`
}

type stakeholder struct {
	Entity db.Entity `json:"entity,omitempty"`
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
	stakeholdersEntities := []db.Entity{}
	for _, stakeholder := range record.Stakeholders {
		stakeholdersEntities = append(stakeholdersEntities, stakeholder.Entity)
	}
	stakeholdersJSON, err := json.Marshal(stakeholdersEntities)
	if err != nil {
		return fmt.Errorf("Failed to marshal sktaeholders: %w", err)
	}
	logger.L.Debugw("Prepared data for insert ISCN", "raw_record", string(data.Record), "record", record, "stakeholders_json", string(stakeholdersJSON), "stakeholders_entities", stakeholdersEntities)
	iscn := db.ISCNInsert{
		Iscn:                 utils.GetEventsValue(events, "iscn_record", "iscn_id"),
		IscnPrefix:           utils.GetEventsValue(events, "iscn_record", "iscn_id_prefix"),
		Version:              getIscnVersion(utils.GetEventsValue(events, "iscn_record", "iscn_id")),
		Owner:                utils.GetEventsValue(events, "iscn_record", "owner"),
		Name:                 record.ContentMetadata.Name,
		Description:          record.ContentMetadata.Description,
		Url:                  record.ContentMetadata.Url,
		Keywords:             utils.ParseKeywords(record.ContentMetadata.Keywords),
		Fingerprints:         record.ContentFingerprints,
		Stakeholders:         stakeholdersJSON,
		StakeholdersEntities: stakeholdersEntities,
		Timestamp:            payload.Timestamp,
		Ipld:                 utils.GetEventsValue(events, "iscn_record", "ipld"),
		Data:                 data.Record,
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
