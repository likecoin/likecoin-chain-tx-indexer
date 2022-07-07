package db

import (
	"log"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

type nftClassMessage struct {
	Input NftClass `json:"input"`
}

type nftClassEvent struct {
	Events types.StringEvents
}

func GetNFTTxs(conn *pgxpool.Conn) {
	ctx, _ := GetTimeoutContext()
	sql := `
	SELECT tx #> '{"tx", "body", "messages"}' AS messages, tx -> 'logs' AS logs
	FROM txs
	WHERE events @> '{"message.action=\"new_class\""}'
	ORDER BY height ASC
	LIMIT 1
	`
	rows, err := conn.Query(ctx, sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	batch := NewBatch(conn, 10)
	for rows.Next() {
		var messageData pgtype.JSONB
		var eventData pgtype.JSONB
		err := rows.Scan(&messageData, &eventData)
		if err != nil {
			panic(err)
		}
		log.Println(string(messageData.Bytes))
		var messages []nftClassMessage
		err = messageData.AssignTo(&messages)
		if err != nil {
			panic(err)
		}
		var events []nftClassEvent
		err = eventData.AssignTo(&events)
		for i, message := range messages {
			event := events[i].Events
			var nftClass NftClass = message.Input
			nftClass.Id = utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "class_id")
			nftClass.Parent = getNftParent(event)
			batch.InsertNFTClass(nftClass)
		}
	}
	err = batch.Flush()
	if err != nil {
		panic(err)
	}
}

func getNftParent(event types.StringEvents) NftClassParent {
	p := NftClassParent{
		IscnIdPrefix: utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "parent_iscn_id_prefix"),
		Account:      utils.GetEventsValue(event, "likechain.likenft.EventNewClass", "parent_account"),
	}
	if p.IscnIdPrefix != "" {
		p.Type = "ISCN"
	} else if p.Account != "" {
		p.Type = "ACCOUNT"
	} else {
		p.Type = "UNKNOWN"
	}
	return p
}

func (batch *Batch) InsertNFTClass(c NftClass) {
	sql := `
	INSERT INTO nft_class
	(id, parent_type, parent_iscn_id_prefix, parent_account,
	name, symbol, description, uri, uri_hash,
	metadata, config, price)
	VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	log.Printf("%+v\n", c)
	batch.Batch.Queue(sql,
		c.Id, c.Parent.Type, c.Parent.IscnIdPrefix, c.Parent.Account,
		c.Name, c.Symbol, c.Description, c.URI, c.URIHash,
		c.Metadata, c.Config, c.Price,
	)
}
