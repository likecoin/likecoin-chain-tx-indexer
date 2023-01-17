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
	ContentFingerprints []string          `json:"contentFingerprints,omitempty"`
	ContentMetadata     *contentMetadata  `json:"contentMetadata,omitempty"`
	Stakeholders        []json.RawMessage `json:"stakeholders,omitempty"`
}

type contentMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Keywords    string `json:"keywords,omitempty"`
}

func (q iscnData) Marshal() ([]byte, error) {
	return json.Marshal(q)
}

func insertIscn(payload db.EventPayload) error {
	message := payload.GetMessage()
	events := payload.GetEvents()
	var data iscnMessage
	if err := json.Unmarshal(message, &data); err != nil {
		return fmt.Errorf("failed to unmarshal iscn: %w", err)
	}
	var record iscnData
	if err := json.Unmarshal(data.Record, &record); err != nil {
		return fmt.Errorf("failed to unmarshal iscn: %w", err)
	}
	stakeholders := []db.Stakeholder{}
	for _, stakeholderRawJSON := range record.Stakeholders {
		parsedStakeholder := db.Stakeholder{}
		err := json.Unmarshal(stakeholderRawJSON, &parsedStakeholder)
		if err != nil {
			logger.L.Errorw("Error when parsing stakeholder JSON", "error", err, "raw_json", string(stakeholderRawJSON))
			// since the fields and sub-fields in db.Stakeholder are set as omitempty, parsing should not fail
			// when it does fail, we leave empty parsedStakeholder as default value, fill in stakeholderRawJSON later
		}
		parsedStakeholder.Data = stakeholderRawJSON
		stakeholders = append(stakeholders, parsedStakeholder)
	}
	iscn := db.IscnInsert{
		Iscn:         utils.GetEventsValue(events, "iscn_record", "iscn_id"),
		IscnPrefix:   utils.GetEventsValue(events, "iscn_record", "iscn_id_prefix"),
		Version:      getIscnVersion(utils.GetEventsValue(events, "iscn_record", "iscn_id")),
		Owner:        utils.GetEventsValue(events, "iscn_record", "owner"),
		Name:         record.ContentMetadata.Name,
		Description:  record.ContentMetadata.Description,
		Url:          record.ContentMetadata.Url,
		Keywords:     utils.ParseKeywords(record.ContentMetadata.Keywords),
		Fingerprints: record.ContentFingerprints,
		Stakeholders: stakeholders,
		Timestamp:    payload.Timestamp,
		Ipld:         utils.GetEventsValue(events, "iscn_record", "ipld"),
		Data:         data.Record,
	}
	payload.Batch.InsertIscn(iscn)
	return nil
}

func transferIscn(payload db.EventPayload) error {
	events := payload.GetEvents()
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
