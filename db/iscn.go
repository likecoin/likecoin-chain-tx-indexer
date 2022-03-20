package db

import (
	"io"
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	iscnTypes "github.com/likecoin/likechain/x/iscn/types"
)

func QueryISCN(conn *pgxpool.Conn, events types.StringEvents) ([]iscnTypes.QueryResponseRecord, error) {
	eventStrings := getEventStrings(events)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	sql := `
		SELECT tx #> '{"tx", "body", "messages", 0, "record"}' as data, events
		FROM txs
		WHERE events @> $1
	`
	rows, err := conn.Query(ctx, sql, eventStrings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseISCNRecords(rows)
}

func parseISCNRecords(rows pgx.Rows) (res []iscnTypes.QueryResponseRecord, err error) {
	res = make([]iscnTypes.QueryResponseRecord, 0)
	for rows.Next() && err == nil {
		var data pgtype.GenericBinary
		var eventsRows pgtype.VarcharArray
		err = rows.Scan(&data, &eventsRows)
		if err != nil {
			return
		}
		var events types.StringEvents
		events, err = parseEvents(eventsRows)
		if err != nil {
			log.Println(err)
		}
		log.Println("events", events)

		result := iscnTypes.QueryResponseRecord{
			Ipld: "",
			Data: iscnTypes.IscnInput(data.Bytes),
		}
		res = append(res, result)
	}
	return
}

func parseEvents(rows pgtype.VarcharArray) (res types.StringEvents, err error) {
	res = make(types.StringEvents, 0)
	for _, row := range rows.Elements {
		var event types.StringEvent
		log.Println(row.String)
		err = event.XXX_Unmarshal([]byte(row.String))
		if err != nil && err != io.ErrUnexpectedEOF {
			panic(err)
		}
		res = append(res, event)
	}
	return
}
