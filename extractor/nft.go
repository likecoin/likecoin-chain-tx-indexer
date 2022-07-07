package extractor

import (
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/db"
)

func extractNFTClass(conn *pgxpool.Conn) (finished bool, err error) {
	begin, err := db.GetMetaHeight(conn, "nft")
	if err != nil {
		return false, fmt.Errorf("Failed to get ISCN synchonized height: %w", err)
	}

	end, err := db.GetLatestHeight(conn)
	if err != nil {
		return false, fmt.Errorf("Failed to get latest height: %w", err)
	}
	if begin == end {
		return true, nil
	}
	if begin+LIMIT < end {
		end = begin + LIMIT
	} else {
		finished = true
	}

	return true, nil

}
