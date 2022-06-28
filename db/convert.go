package db

import (
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
)

type ISCN struct {
	Id           uint64
	Iscn         string
	Owner        string
	Keywords     []string
	Fingerprints []string
	Stakeholders []byte
	Data         []byte
}

func ConvertISCN(pool *pgxpool.Pool, pagination Pagination) error {
	log.Println("Converting")
	conn, err := AcquireFromPool(pool)
	if err != nil {
		return err
	}
	defer conn.Release()

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := `
		SELECT id, tx #> '{"tx", "body", "messages", 0, "record"}' as records, events, tx #> '{"timestamp"}'
		FROM txs
		WHERE events && '{"message.module=\"iscn\""}'
		ORDER BY id ASC
		OFFSET $1
		LIMIT $2;
	`
	rows, err := conn.Query(ctx, sql, pagination.getOffset(), pagination.Limit)
	if err != nil {
		logger.L.Errorw("Query error:", "error", err)
		return err
	}
	defer rows.Close()
	return handleISCNRecords(rows)
}

func handleISCNRecords(rows pgx.Rows) (err error) {
	for rows.Next() {
		var id uint64
		var data pgtype.JSONB
		var eventsRows pgtype.VarcharArray
		var timestamp string
		err = rows.Scan(&id, &data, &eventsRows, &timestamp)
		if err != nil {
			return
		}

		log.Println(id)
		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			logger.L.Warnw("Failed to parse events of db rows", "id", id, "error", err)
			return
		}

		switch getEventsValue(events, "message", "action") {
		case "create_iscn_record", "/likechain.iscn.MsgCreateIscnRecord":
			log.Println("create")
		case "update_iscn_record", "/likechain.iscn.MsgUpdateIscnRecord":
			log.Println("update")
		case "msg_change_iscn_record_ownership", "/likechain.iscn.MsgChangeIscnRecordOwnership":
			log.Println("transfer")
		default:
			log.Println("other:", getEventsValue(events, "message", "action"))
		}
	}
	return
}
