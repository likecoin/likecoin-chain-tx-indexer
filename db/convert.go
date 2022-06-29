package db

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

type ISCN struct {
	Iscn         string
	Owner        string
	Keywords     []string
	Fingerprints []string
	Stakeholders []byte
	Data         []byte
}

func ConvertISCN(conn *pgxpool.Conn, limit int) error {
	log.Println("Converting", limit)

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	height, err := getHeight(conn)
	if err != nil {
		log.Fatalln(err)
	}

	maxHeight, err := GetLatestHeight(conn)
	if err != nil {
		log.Fatalln(err)
	}
	if height+int64(limit) < maxHeight {
		maxHeight = height + int64(limit)
	}

	log.Println("Previous height:", height)

	sql := `
		SELECT height, tx #> '{"tx", "body", "messages", 0, "record"}' as record, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE height >= $1
			AND height < $2
			AND events @> '{"message.module=\"iscn\""}'
		ORDER BY height ASC;
	`
	rows, err := conn.Query(ctx, sql, height, maxHeight)
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return err
	}
	defer rows.Close()
	batch := NewBatch(conn, int(limit))
	batch.prevHeight = maxHeight

	return handleISCNRecords(rows, &batch)
}

func getHeight(conn *pgxpool.Conn) (int64, error) {
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	row := conn.QueryRow(ctx, `SELECT height FROM meta WHERE id = 'iscn'`)
	var height int64
	err := row.Scan(&height)
	return height, err
}

func handleISCNRecords(rows pgx.Rows, batch *Batch) (err error) {
	for rows.Next() {
		var height int64
		var data pgtype.JSONB
		var eventsRows pgtype.VarcharArray
		var timestamp string
		err = rows.Scan(&height, &data, &eventsRows, &timestamp)
		if err != nil {
			return
		}

		log.Println(height)
		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			logger.L.Warnw("Failed to parse events of db rows", "height", height, "error", err)
			return
		}

		switch getEventsValue(events, "message", "action") {
		case "create_iscn_record", "/likechain.iscn.MsgCreateIscnRecord":
			log.Println("create")
			batch.InsertISCN(events, data, timestamp)
		case "update_iscn_record", "/likechain.iscn.MsgUpdateIscnRecord":
			log.Println("update")
		case "msg_change_iscn_record_ownership", "/likechain.iscn.MsgChangeIscnRecordOwnership":
			log.Println("transfer")
		default:
			log.Println("other:", getEventsValue(events, "message", "action"))
		}
	}
	batch.Batch.Queue(`UPDATE meta SET height = $1 WHERE id = 'iscn'`, batch.prevHeight)
	log.Println("Last height:", batch.prevHeight)
	return batch.Flush()
}

func parseKeywords(keyword string) []string {
	arr := strings.Split(keyword, ",")
	for i, v := range arr {
		arr[i] = strings.TrimSpace(v)
	}
	return arr
}

func formatStakeholders(stakeholders []Stakeholder) ([]byte, error) {
	body := make([]*Entity, len(stakeholders))
	for i, v := range stakeholders {
		body[i] = v.Entity
	}
	return json.Marshal(body)
}

func (batch *Batch) InsertISCN(events types.StringEvents, data pgtype.JSONB, timestamp string) {
	var record ISCNRecordQuery
	if err := json.Unmarshal(data.Bytes, &record); err != nil {
		log.Fatalln(err)
	}
	holders, err := formatStakeholders(record.Stakeholders)
	if err != nil {
		log.Fatalln(err)
	}
	iscn := ISCN{
		Iscn:         getEventsValue(events, "iscn_record", "iscn_id"),
		Owner:        getEventsValue(events, "iscn_record", "owner"),
		Keywords:     parseKeywords(record.ContentMetadata.Keywords),
		Fingerprints: record.ContentFingerprints,
		Stakeholders: holders,
		Data:         data.Bytes,
	}
	sql := `
INSERT INTO iscn (iscn_id, owner, keywords, fingerprints, stakeholders, data) VALUES
( $1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING;`
	batch.Batch.Queue(sql, iscn.Iscn, iscn.Owner, iscn.Keywords, iscn.Fingerprints, iscn.Stakeholders, iscn.Data)
}
