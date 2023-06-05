package db

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/likecoin/likecoin-chain-tx-indexer/logger"
	"github.com/likecoin/likecoin-chain-tx-indexer/utils"
)

func GetClasses(conn *pgxpool.Conn, q QueryClassRequest, p PageRequest) (QueryClassResponse, error) {
	accountVariations := utils.ConvertAddressPrefixes(q.Account, AddressPrefixes)
	iscnOwnerVariations := utils.ConvertAddressArrayPrefixes(q.IscnOwner, AddressPrefixes)
	sql := fmt.Sprintf(`
	SELECT DISTINCT ON (c.id)
		c.id, c.class_id, c.name, c.description, c.symbol,
		c.uri, c.uri_hash, c.config, c.metadata, c.latest_price,
		c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at, c.price_updated_at, i.owner,
	(
		SELECT array_agg(row_to_json((n.*)))
		FROM nft as n
		WHERE n.class_id = c.class_id
			AND $7 = true
		GROUP BY n.class_id
	) as nfts
	FROM nft_class as c
	LEFT JOIN iscn AS i ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	LEFT JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
	WHERE ($4 = '' OR c.parent_iscn_id_prefix = $4)
		AND ($5::text[] IS NULL OR cardinality($5::text[]) = 0 OR c.parent_account = ANY($5))
		AND ($6::text[] IS NULL OR cardinality($6::text[]) = 0 OR i.owner = ANY($6))
		AND ($8 = true OR i.version = iscn_latest_version.latest_version)
		AND ($1 = 0 OR c.id > $1)
		AND ($2 = 0 OR c.id < $2)
	ORDER BY c.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(
		ctx, sql,
		p.After(), p.Before(), p.Limit, q.IscnIdPrefix, accountVariations,
		iscnOwnerVariations, q.Expand, q.AllIscnVersions)
	if err != nil {
		logger.L.Errorw("Failed to query nft class by iscn id prefix", "error", err, "q", q)
		return QueryClassResponse{}, fmt.Errorf("query nft class by iscn id prefix error: %w", err)
	}

	res := QueryClassResponse{
		Classes: make([]NftClassResponse, 0),
	}
	for rows.Next() {
		var c NftClassResponse
		var nfts pgtype.JSONBArray
		if err = rows.Scan(
			&res.Pagination.NextKey, &c.Id, &c.Name, &c.Description, &c.Symbol,
			&c.URI, &c.URIHash, &c.Config, &c.Metadata, &c.LatestPrice,
			&c.Parent.Type, &c.Parent.IscnIdPrefix, &c.Parent.Account, &c.CreatedAt, &c.PriceUpdatedAt,
			&c.Owner, &nfts,
		); err != nil {
			logger.L.Errorw("failed to scan nft class", "error", err)
			return QueryClassResponse{}, fmt.Errorf("query nft class data failed: %w", err)
		}
		if err = nfts.AssignTo(&c.Nfts); err != nil {
			logger.L.Errorw("failed to scan nfts", "error", err)
			return QueryClassResponse{}, fmt.Errorf("query nfts failed: %w", err)
		}
		c.Count = len(c.Nfts)
		res.Classes = append(res.Classes, c)
	}
	res.Pagination.Count = len(res.Classes)
	return res, nil
}

func GetClassesRanking(conn *pgxpool.Conn, q QueryRankingRequest, p PageRequest) (QueryRankingResponse, error) {
	stakeholderIdVariataions := utils.ConvertAddressPrefixes(q.StakeholderId, AddressPrefixes)
	creatorVariations := utils.ConvertAddressPrefixes(q.Creator, AddressPrefixes)
	collectorVariations := utils.ConvertAddressPrefixes(q.Collector, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	ApiAddressesVariations := utils.ConvertAddressArrayPrefixes(q.ApiAddresses, AddressPrefixes)
	orderBy := q.OrderBy
	switch orderBy {
	case "total_sold_value", "sold_count":
	default:
		orderBy = "total_sold_value"
	}
	sql := fmt.Sprintf(`
	SELECT
		c.class_id, c.name, c.description, c.symbol, c.uri,
		c.uri_hash, c.config, c.metadata, c.latest_price, c.parent_type,
		c.parent_iscn_id_prefix, c.parent_account, c.created_at, c.price_updated_at,
		COUNT(DISTINCT t.nft_id) AS sold_count,
		SUM(t.price) AS total_sold_value
	FROM (
		WITH ne AS (
			SELECT
				n.id, n.nft_id, n.class_id, n.owner, e.price
			FROM nft AS n
			JOIN nft_event AS e
			ON n.class_id = e.class_id
				AND n.nft_id = e.nft_id
			WHERE
				($3::text[] IS NULL OR cardinality($3::text[]) = 0 OR n.owner != ALL($3))
				AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR n.owner = ANY($8))
				AND e.action = '/cosmos.nft.v1beta1.MsgSend'
				AND ($11 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp > to_timestamp($11)))
				AND ($12 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp < to_timestamp($12)))
				AND e.sender = ANY($13::text[])
		)
		SELECT DISTINCT ON (ne.id)
			ne.nft_id,
			c.id AS class_pid,
			ne.price AS price
		FROM ne
		JOIN nft_class AS c
			ON c.class_id = ne.class_id
		JOIN iscn AS i
			ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN iscn_latest_version
			ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
				AND i.version = iscn_latest_version.latest_version
		LEFT JOIN iscn_stakeholders
			ON (
				-- this is for optimizing out a left join when stakeholder data is not needed
				(
					($6::text[] IS NOT NULL AND cardinality($6::text[]) > 0)
					OR $7 != ''
				)
				AND i.id = iscn_pid
			)
		WHERE
			($2 = true OR ne.owner != i.owner)
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR i.owner = ANY($4))
			AND ($5 = '' OR i.data #>> '{"contentMetadata", "@type"}' = $5)
			AND ($6::text[] IS NULL OR cardinality($6::text[]) = 0 OR sid = ANY($6))
			AND ($7 = '' OR sname = $7)
			AND ($9 = 0 OR c.created_at > to_timestamp($9))
			AND ($10 = 0 OR c.created_at < to_timestamp($10))
	) AS t
	JOIN nft_class AS c
		ON c.id = t.class_pid
	GROUP BY c.id
	ORDER BY %[1]s DESC
	LIMIT $1
	`, orderBy)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		// $1 ~ $5
		p.Limit, q.IncludeOwner, ignoreListVariations, creatorVariations, q.Type,
		// $6 ~ $10
		stakeholderIdVariataions, q.StakeholderName, collectorVariations, q.CreatedAfter, q.CreatedBefore,
		// $11 ~ $13
		q.After, q.Before, ApiAddressesVariations,
	)
	if err != nil {
		logger.L.Errorw("Failed to query nft class ranking", "error", err, "q", q)
		return QueryRankingResponse{}, fmt.Errorf("query nft class ranking error: %w", err)
	}

	res := QueryRankingResponse{}
	for rows.Next() {
		var c NftClassRankingResponse
		if err = rows.Scan(
			&c.Id, &c.Name, &c.Description, &c.Symbol, &c.URI,
			&c.URIHash, &c.Config, &c.Metadata, &c.LatestPrice, &c.Parent.Type,
			&c.Parent.IscnIdPrefix, &c.Parent.Account, &c.CreatedAt, &c.PriceUpdatedAt,
			&c.SoldCount, &c.TotalSoldValue,
		); err != nil {
			logger.L.Errorw("failed to scan nft class", "error", err)
			return QueryRankingResponse{}, fmt.Errorf("query nft class data failed: %w", err)
		}
		res.Classes = append(res.Classes, c)
	}
	res.Pagination.Count = len(res.Classes)
	return res, nil
}

func GetNfts(conn *pgxpool.Conn, q QueryNftRequest, p PageRequest) (QueryNftResponse, error) {
	ownerVariations := utils.ConvertAddressPrefixes(q.Owner, AddressPrefixes)
	sql := fmt.Sprintf(`
	SELECT
		n.id, n.nft_id, n.class_id, n.owner, n.uri,
		n.uri_hash, n.metadata, e.timestamp, c.name, c.description,
		c.symbol, c.uri, c.uri_hash, c.config, c.metadata,
		c.latest_price, c.parent_type, c.parent_iscn_id_prefix, c.parent_account, c.created_at,
		c.price_updated_at
	FROM nft as n
	JOIN nft_class as c
	ON n.class_id = c.class_id
	JOIN (
		SELECT DISTINCT ON (nft_id) nft_id, timestamp
		FROM nft_event
		WHERE receiver = ANY($4)
		ORDER BY nft_id, timestamp DESC
	) e
	ON n.nft_id = e.nft_id
	WHERE owner = ANY($4)
		AND ($1 = 0 OR n.id > $1)
		AND ($2 = 0 OR n.id < $2)
	ORDER BY n.id %s
	LIMIT $3
	`, p.Order())
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, p.After(), p.Before(), p.Limit, ownerVariations)
	if err != nil {
		logger.L.Errorw("Failed to query nft by owner", "error", err, "q", q)
		return QueryNftResponse{}, fmt.Errorf("query nft class error: %w", err)
	}
	res := QueryNftResponse{
		Nfts: make([]NftResponse, 0),
	}
	for rows.Next() {
		var n NftResponse
		var c NftClass
		if err = rows.Scan(
			&res.Pagination.NextKey, &n.NftId, &n.ClassId, &n.Owner, &n.Uri,
			&n.UriHash, &n.Metadata, &n.Timestamp, &c.Name, &c.Description,
			&c.Symbol, &c.URI, &c.URIHash, &c.Config, &c.Metadata,
			&c.LatestPrice, &n.ClassParent.Type, &n.ClassParent.IscnIdPrefix, &n.ClassParent.Account, &c.CreatedAt,
			&c.PriceUpdatedAt,
		); err != nil {
			logger.L.Errorw("failed to scan nft", "error", err, "q", q)
			return QueryNftResponse{}, fmt.Errorf("query nft failed: %w", err)
		}
		if q.ExpandClasses {
			c.Parent = n.ClassParent
			c.Id = n.ClassId
			n.ClassData = &c
		}
		res.Nfts = append(res.Nfts, n)
	}
	res.Pagination.Count = len(res.Nfts)
	return res, nil
}

func GetOwners(conn *pgxpool.Conn, q QueryOwnerRequest) (QueryOwnerResponse, error) {
	sql := `
	SELECT i.owner FROM iscn AS i 
	JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND i.version = iscn_latest_version.latest_version
	JOIN nft_class AS c
		ON iscn_latest_version.iscn_id_prefix = c.parent_iscn_id_prefix
	WHERE c.class_id = $1
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql, q.ClassId)
	if err != nil {
		logger.L.Errorw("Failed to query class owner", "error", err)
		return QueryOwnerResponse{}, fmt.Errorf("query owner error: %w", err)
	}
	var owner string
	for rows.Next() {
		if err = rows.Scan(&owner); err != nil {
			logger.L.Errorw("failed to scan owner", "error", err, "q", q)
			return QueryOwnerResponse{}, fmt.Errorf("query owner data failed: %w", err)
		}
	}

	sql = `
	SELECT owner, array_agg(nft_id)
	FROM nft
	WHERE class_id = $1
	GROUP BY owner
	`

	rows, err = conn.Query(ctx, sql, q.ClassId)
	if err != nil {
		logger.L.Errorw("Failed to query owner", "error", err)
		return QueryOwnerResponse{}, fmt.Errorf("query owner error: %w", err)
	}

	res := QueryOwnerResponse{
		Owners: make([]OwnerResponse, 0),
	}
	for rows.Next() {
		var ownerRes OwnerResponse
		if err = rows.Scan(&ownerRes.Owner, &ownerRes.Nfts); err != nil {
			logger.L.Errorw("failed to scan owner response", "error", err, "q", q)
			return QueryOwnerResponse{}, fmt.Errorf("query owner data failed: %w", err)
		}
		if ownerRes.Owner != owner {
			res.OwnerCountExcludeIscnOwner++
			res.CollectedCountExcludeIscnOwner += len(ownerRes.Nfts)
		}
		ownerRes.Count = len(ownerRes.Nfts)
		res.Owners = append(res.Owners, ownerRes)
	}
	res.Pagination.Count = len(res.Owners)
	return res, nil
}

func GetNftEvents(conn *pgxpool.Conn, q QueryEventsRequest, p PageRequest) (QueryEventsResponse, error) {
	ignoreFromListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreFromList, AddressPrefixes)
	ignoreToListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreToList, AddressPrefixes)
	senderVariations := utils.ConvertAddressArrayPrefixes(q.Sender, AddressPrefixes)
	receiverVariations := utils.ConvertAddressArrayPrefixes(q.Receiver, AddressPrefixes)
	creatorVariations := utils.ConvertAddressArrayPrefixes(q.Creator, AddressPrefixes)
	involverVariations := utils.ConvertAddressArrayPrefixes(q.Involver, AddressPrefixes)
	sql := fmt.Sprintf(`
		SELECT DISTINCT ON (e.id)
			e.id, e.action, e.class_id, e.nft_id, e.sender,
			e.receiver, e.timestamp, e.tx_hash, e.events, e.price,
			e.memo
		FROM (
			(
				SELECT DISTINCT ON (e.id) e.*
				FROM nft_event as e
				JOIN nft_class as c
				ON e.class_id = c.class_id
				JOIN iscn AS i
				ON i.iscn_id_prefix = c.parent_iscn_id_prefix
				JOIN iscn_latest_version
				ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
					AND i.version = iscn_latest_version.latest_version
				WHERE ($4 = '' OR e.class_id = $4)
					AND ($12::text[] IS NULL OR cardinality($12::text[]) = 0 OR i.owner = ANY($12))
					AND (nft_id = '' OR $5 = '' OR nft_id = $5)
					AND ($6 = '' OR c.parent_iscn_id_prefix = $6)
					AND ($10::text[] IS NULL OR cardinality($10::text[]) = 0 OR e.sender = ANY($10))
					AND ($11::text[] IS NULL OR cardinality($11::text[]) = 0 OR e.receiver = ANY($11))
					AND ($13::text[] IS NULL OR cardinality($13::text[]) = 0
						OR e.sender = ANY($13)
						OR e.receiver = ANY($13)
					)
					AND ($1 = 0 OR e.id > $1)
					AND ($2 = 0 OR e.id < $2)
					AND ($7::text[] IS NULL OR cardinality($7::text[]) = 0 OR e.action = ANY($7))
					AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR e.sender != ALL($8))
					AND ($9::text[] IS NULL OR cardinality($9::text[]) = 0 OR e.receiver != ALL($9))
				ORDER BY e.id %[1]s
				LIMIT $3
			) UNION ALL (
				SELECT DISTINCT ON (e.id) e.*
				FROM nft_event as e
				JOIN nft_class as c
				ON e.class_id = c.class_id
				JOIN iscn AS i
				ON i.iscn_id_prefix = c.parent_iscn_id_prefix
				JOIN iscn_latest_version
				ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
					AND i.version = iscn_latest_version.latest_version
				WHERE ($4 = '' OR e.class_id = $4)
					AND ($12::text[] IS NULL OR cardinality($12::text[]) = 0 OR i.owner = ANY($12))
					AND (nft_id = '' OR $5 = '' OR nft_id = $5)
					AND ($6 = '' OR c.parent_iscn_id_prefix = $6)
					AND ($10::text[] IS NULL OR cardinality($10::text[]) = 0 OR e.sender = ANY($10))
					AND ($11::text[] IS NULL OR cardinality($11::text[]) = 0 OR e.receiver = ANY($11))
					AND ($13::text[] IS NULL OR cardinality($13::text[]) = 0 OR i.owner = ANY($13))
					AND ($1 = 0 OR e.id > $1)
					AND ($2 = 0 OR e.id < $2)
					AND ($7::text[] IS NULL OR cardinality($7::text[]) = 0 OR e.action = ANY($7))
					AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR e.sender != ALL($8))
					AND ($9::text[] IS NULL OR cardinality($9::text[]) = 0 OR e.receiver != ALL($9))
				ORDER BY e.id %[1]s
				LIMIT $3
			)
		) AS e
		ORDER BY e.id %[1]s
		LIMIT $3
	`, p.Order())

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(
		ctx, sql,
		p.After(), p.Before(), p.Limit, q.ClassId, q.NftId,
		q.IscnIdPrefix, q.ActionType, ignoreFromListVariations, ignoreToListVariations, senderVariations,
		receiverVariations, creatorVariations, involverVariations,
	)
	if err != nil {
		logger.L.Errorw("Failed to query nft events", "error", err)
		return QueryEventsResponse{}, fmt.Errorf("query nft events error: %w", err)
	}

	res := QueryEventsResponse{
		Events: make([]NftEvent, 0),
	}
	for rows.Next() {
		var e NftEvent
		var eventRaw []string
		var price *uint64
		if err = rows.Scan(
			&res.Pagination.NextKey, &e.Action, &e.ClassId, &e.NftId, &e.Sender,
			&e.Receiver, &e.Timestamp, &e.TxHash, &eventRaw, &price,
			&e.Memo,
		); err != nil {
			logger.L.Errorw("failed to scan nft events", "error", err, "q", q)
			return QueryEventsResponse{}, fmt.Errorf("query nft events data failed: %w", err)
		}
		if price != nil {
			e.Price = *price
		}
		if q.Verbose {
			e.Events, err = utils.ParseEvents(eventRaw)
			if err != nil {
				logger.L.Errorw("failed to parse events", "error", err, "event_raw", eventRaw)
				return QueryEventsResponse{}, fmt.Errorf("parse nft events data failed: %w", err)
			}
		}
		res.Events = append(res.Events, e)
	}
	res.Pagination.Count = len(res.Events)
	return res, nil
}

func GetNftIncomes(conn *pgxpool.Conn, q QueryIncomesRequest, p PageRequest) (QueryIncomesResponse, error) {
	ownerVariations := utils.ConvertAddressPrefixes(q.Owner, AddressPrefixes)
	beneficiaryVariations := utils.ConvertAddressPrefixes(q.Address, AddressPrefixes)

	ownershipCondition := "true"
	switch q.IscnOwnership {
	case "owned":
		ownershipCondition = "i.address = e.iscn_owner_at_the_time"
	case "not_owned":
		ownershipCondition = "i.address != e.iscn_owner_at_the_time"
	case "all":
	}

	orderBy := "class_income"
	switch q.OrderBy {
	case "created_time":
		orderBy = "created_at"
	case "income":
	}

	sql := fmt.Sprintf(`
		SELECT * FROM (
			SELECT *, DENSE_RANK() OVER(ORDER BY %[1]s DESC, class_id) as class_rank
			FROM (
				SELECT e.class_id, c.created_at, i.address, 
					SUM(i.amount) AS income,
					SUM(SUM(i.amount)) OVER(PARTITION BY e.class_id) as class_income
				FROM nft_event AS e
				JOIN nft_class AS c
					ON e.class_id = c.class_id
				JOIN nft_income AS i
					ON e.class_id = i.class_id
					AND e.nft_id = i.nft_id
					AND e.tx_hash = i.tx_hash
				WHERE e.price > 0
					AND ($3 = '' OR e.class_id = $3)
					AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR e.iscn_owner_at_the_time = ANY($4))
					AND ($5::text[] IS NULL OR cardinality($5::text[]) = 0 OR i.address = ANY($5))
					AND ($6 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp > to_timestamp($6)))
					AND ($7 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp < to_timestamp($7)))
					AND ($8::text[] IS NULL OR cardinality($8::text[]) = 0 OR e.action = ANY($8))
					AND ($9 = false OR e.receiver != e.iscn_owner_at_the_time)
					AND (%[2]s)
				GROUP BY e.class_id, c.created_at, i.address
			) AS sum_query
			ORDER BY class_rank, income DESC
		) AS rank_query
		WHERE class_rank > $1 AND class_rank <= $2
	`, orderBy, ownershipCondition)

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rankStart := int(p.Key)
	rankEnd := rankStart + p.Limit
	rows, err := conn.Query(
		ctx, sql,
		rankStart, rankEnd, q.ClassId, ownerVariations, beneficiaryVariations,
		q.After, q.Before, q.ActionType, q.ExcludeSelfPurchase,
	)
	if err != nil {
		logger.L.Errorw("Failed to query nft incomes", "error", err)
		return QueryIncomesResponse{}, fmt.Errorf("query nft incomes error: %w", err)
	}

	res := QueryIncomesResponse{
		ClassIncomes: make([]NftClassIncomeResponse, 0),
	}

	var rank int
	rankMap := make(map[int]NftClassIncomeResponse)
	for rows.Next() {
		var ci NftClassIncomeResponse
		var i NftIncomeResponse

		if err = rows.Scan(&ci.ClassId, &ci.CreatedAt, &i.Address, &i.Amount, &ci.TotalAmount, &rank); err != nil {
			logger.L.Errorw("failed to scan nft incomes", "error", err, "q", q)
			return QueryIncomesResponse{}, fmt.Errorf("query nft incomes data failed: %w", err)
		}

		classIncome, exists := rankMap[rank]
		if !exists {
			classIncome = ci
			classIncome.Incomes = make([]NftIncomeResponse, 0)
		}

		classIncome.Incomes = append(classIncome.Incomes, i)
		rankMap[rank] = classIncome
		res.TotalAmount += i.Amount
	}
	var classIds []string
	for _, classIncome := range rankMap {
		res.ClassIncomes = append(res.ClassIncomes, classIncome)
		classIds = append(classIds, classIncome.ClassId)
	}
	for i := 0; i < len(res.ClassIncomes); i++ {
		res.ClassIncomes[i] = rankMap[rankStart+i+1]
	}
	res.Pagination.NextKey = uint64(rank)
	res.Pagination.Count = len(res.ClassIncomes)

	sql = `
		SELECT e.class_id, SUM(e.price) AS sales
		FROM nft_event AS e
		WHERE e.class_id = ANY($1)
			AND ($2 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp > to_timestamp($2)))
			AND ($3 = 0 OR (e.timestamp IS NOT NULL AND e.timestamp < to_timestamp($3)))
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR e.action = ANY($4))
			AND ($5 = false OR e.receiver != e.iscn_owner_at_the_time)
		GROUP BY e.class_id
	`
	rows, err = conn.Query(
		ctx, sql,
		classIds, q.After, q.Before, q.ActionType, q.ExcludeSelfPurchase,
	)
	if err != nil {
		logger.L.Errorw("Failed to query nft sales", "error", err)
		return QueryIncomesResponse{}, fmt.Errorf("query nft sales error: %w", err)
	}

	salesMap := make(map[string]uint64)
	for rows.Next() {
		var classId string
		var sales uint64

		if err = rows.Scan(&classId, &sales); err != nil {
			logger.L.Errorw("failed to scan nft sales", "error", err, "q", q)
			return QueryIncomesResponse{}, fmt.Errorf("query nft sales data failed: %w", err)
		}

		salesMap[classId] = sales
	}

	for i := range res.ClassIncomes {
		classId := res.ClassIncomes[i].ClassId
		sales := salesMap[classId]
		res.ClassIncomes[i].Sales = sales
		res.TotalSales += sales
	}

	return res, nil
}

func getTotalValueSourceField(priceBy string) string {
	switch priceBy {
	case "class":
		return "c.latest_price"
	case "nft":
	default:
		return "e.price"
	}
	return "e.price"
}

func convertOrderBy(orderBy string) string {
	switch orderBy {
	case "count":
		return "total_count"
	case "price":
	default:
		return "total_value"
	}
	return "total_value"
}

func GetCollector(conn *pgxpool.Conn, q QueryCollectorRequest, p PageRequest) (res QueryCollectorResponse, err error) {
	creatorVariations := utils.ConvertAddressPrefixes(q.Creator, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	totalValueSourceField := getTotalValueSourceField(q.PriceBy)
	orderBy := convertOrderBy(q.OrderBy)
	sql := fmt.Sprintf(`
	SELECT owner, SUM(value) AS total_value, SUM(count) AS total_count,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'value', value,
			'count', count)),
		COUNT(*) OVER() AS row_count
	FROM (
		SELECT n.owner, i.iscn_id_prefix, c.class_id, SUM(%[1]s) AS value, COUNT(DISTINCT n.id) as count
		FROM iscn AS i
		JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND ($5 = true OR i.version = iscn_latest_version.latest_version)
		JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR n.owner != ALL($4))
		JOIN LATERAL (
			SELECT nft_id, receiver, MAX(id) AS max_id
			FROM nft_event
			WHERE price IS NOT NULL
				AND nft_id = n.nft_id
				AND receiver = n.owner
			GROUP BY nft_id, receiver
		) AS latest_e 
		ON latest_e.nft_id = n.nft_id
			AND latest_e.receiver = n.owner
		JOIN nft_event AS e 
		ON e.nft_id = n.nft_id
			AND e.receiver = n.owner
			AND e.id = latest_e.max_id
		WHERE 
			($6 = true OR n.owner != i.owner)
			AND ($1::text[] IS NULL OR cardinality($1::text[]) = 0 OR i.owner = ANY($1))
		GROUP BY n.owner, i.iscn_id_prefix, c.class_id
	) AS r
	GROUP BY owner
	ORDER BY %[2]s DESC, owner DESC
	OFFSET $2
	LIMIT $3
	`, totalValueSourceField, orderBy)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		creatorVariations, p.Offset, p.Limit, ignoreListVariations, q.AllIscnVersions,
		q.IncludeOwner)
	if err != nil {
		logger.L.Errorw("failed to query collectors", "error", err, "q", q)
		err = fmt.Errorf("query supporters error: %w", err)
		return
	}
	defer rows.Close()

	res.Collectors, res.Pagination.Total, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("scan collectors error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Collectors)
	return
}

func GetCreators(conn *pgxpool.Conn, q QueryCreatorRequest, p PageRequest) (res QueryCreatorResponse, err error) {
	collectorVariations := utils.ConvertAddressPrefixes(q.Collector, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	totalValueSourceField := getTotalValueSourceField(q.PriceBy)
	orderBy := convertOrderBy(q.OrderBy)
	sql := fmt.Sprintf(`
	SELECT owner, SUM(value) as total_value, SUM(count) AS total_count,
		array_agg(json_build_object(
			'iscn_id_prefix', iscn_id_prefix,
			'class_id', class_id,
			'value', value,
			'count', count)),
		COUNT(*) OVER() AS row_count
	FROM (
		SELECT i.owner, i.iscn_id_prefix, c.class_id, SUM(%[1]s) AS value, COUNT(DISTINCT n.id) as count
		FROM iscn AS i
		JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND ($5 = true OR i.version = iscn_latest_version.latest_version)
		JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
			AND ($4::text[] IS NULL OR cardinality($4::text[]) = 0 OR n.owner != ALL($4))
		JOIN LATERAL (
			SELECT nft_id, receiver, MAX(id) AS max_id
			FROM nft_event
			WHERE price IS NOT NULL
				AND nft_id = n.nft_id
				AND receiver = n.owner
			GROUP BY nft_id, receiver
		) AS latest_e 
		ON latest_e.nft_id = n.nft_id
			AND latest_e.receiver = n.owner
		JOIN nft_event AS e 
		ON e.nft_id = n.nft_id
			AND e.receiver = n.owner
			AND e.id = latest_e.max_id
		WHERE 
			($6 = true OR n.owner != i.owner)
			AND ($1::text[] IS NULL OR cardinality($1::text[]) = 0 OR n.owner = ANY($1))
		GROUP BY i.owner, i.iscn_id_prefix, c.class_id
	) AS r
	GROUP BY owner
	ORDER BY %[2]s DESC
	OFFSET $2
	LIMIT $3
	`, totalValueSourceField, orderBy)
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		collectorVariations, p.Offset, p.Limit, ignoreListVariations, q.AllIscnVersions,
		q.IncludeOwner)
	if err != nil {
		logger.L.Errorw("failed to query creators", "error", err, "q", q)
		err = fmt.Errorf("query creators error: %w", err)
		return
	}

	res.Creators, res.Pagination.Total, err = parseAccountCollections(rows)
	if err != nil {
		err = fmt.Errorf("scan creators error: %w", err)
		return
	}
	res.Pagination.Count = len(res.Creators)
	return
}

func parseAccountCollections(rows pgx.Rows) (accounts []accountCollection, rowCount int, err error) {
	for rows.Next() {
		var account accountCollection
		var collections pgtype.JSONBArray
		if err = rows.Scan(&account.Account, &account.TotalValue, &account.Count, &collections, &rowCount); err != nil {
			return
		}
		if err = collections.AssignTo(&account.Collections); err != nil {
			return
		}
		accounts = append(accounts, account)
	}
	return
}

func GetUserStat(conn *pgxpool.Conn, q QueryUserStatRequest) (res QueryUserStatResponse, err error) {
	res = QueryUserStatResponse{
		CollectedClasses: make([]CollectedClass, 0),
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()

	sql := `
	SELECT c.class_id, COUNT(c.id)
	FROM nft_class as c
	JOIN nft AS n ON c.class_id = n.class_id
	WHERE n.owner = $1
	GROUP BY c.class_id
	`
	rows, err := conn.Query(ctx, sql, q.User)
	if err != nil {
		logger.L.Errorw("failed to query collected classes", "error", err, "q", q)
		err = fmt.Errorf("query collected classes error: %w", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var c CollectedClass
		if err = rows.Scan(&c.ClassId, &c.Count); err != nil {
			err = fmt.Errorf("scan collected classes error: %w", err)
			return
		}
		res.CollectedClasses = append(res.CollectedClasses, c)
	}

	sql = `
	SELECT COUNT(DISTINCT(c.class_id))
	FROM iscn AS i
	JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
		AND ($2 = true OR i.version = iscn_latest_version.latest_version)
	JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	WHERE i.owner = $1
	`

	row := conn.QueryRow(ctx, sql, q.User, q.AllIscnVersions)

	if err = row.Scan(&res.CreatedCount); err != nil {
		err = fmt.Errorf("scan created count error: %w", err)
		return
	}

	sql = `
	SELECT COUNT(DISTINCT(n.owner))
	FROM iscn AS i
	JOIN iscn_latest_version
	ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
		AND ($3 = true OR i.version = iscn_latest_version.latest_version)
	JOIN nft_class AS c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
	JOIN nft AS n ON c.class_id = n.class_id
		AND ($2::text[] IS NULL OR n.owner != ALL($2))
	WHERE i.owner = $1
	`

	row = conn.QueryRow(ctx, sql, q.User, q.IgnoreList, q.AllIscnVersions)

	err = row.Scan(&res.CollectorCount)
	if err != nil {
		err = fmt.Errorf("scan collector count error: %w", err)
		return
	}

	sql = `
	SELECT COALESCE(SUM(e.price), 0)
	FROM nft_event AS e
	WHERE e.iscn_owner_at_the_time = $1
		AND e.price IS NOT NULL
	`

	row = conn.QueryRow(ctx, sql, q.User)

	err = row.Scan(&res.TotalSales)
	if err != nil {
		err = fmt.Errorf("scan total sales error: %w", err)
		return
	}

	sql = `
	SELECT COALESCE(SUM(amount), 0)
	FROM nft_income
	WHERE address = $1
	`

	row = conn.QueryRow(ctx, sql, q.User)

	err = row.Scan(&res.TotalIncomes)
	if err != nil {
		err = fmt.Errorf("scan total income error: %w", err)
		return
	}

	return
}

func GetCollectorTopRankedCreators(conn *pgxpool.Conn, q QueryCollectorTopRankedCreatorsRequest) (res QueryCollectorTopRankedCreatorsResponse, err error) {
	collectorVariations := utils.ConvertAddressPrefixes(q.Collector, AddressPrefixes)
	ignoreListVariations := utils.ConvertAddressArrayPrefixes(q.IgnoreList, AddressPrefixes)
	sql := `
	SELECT creator, rank FROM (
		SELECT
			i.owner AS creator,
			n.owner AS collector,
			SUM(n.latest_price) AS total_value,
			RANK() OVER (PARTITION BY i.owner ORDER BY SUM(n.latest_price) DESC) AS rank
		FROM iscn as i
		JOIN iscn_latest_version
		ON i.iscn_id_prefix = iscn_latest_version.iscn_id_prefix
			AND ($4 = true OR i.version = iscn_latest_version.latest_version)
		JOIN nft_class as c ON i.iscn_id_prefix = c.parent_iscn_id_prefix
		JOIN nft AS n ON c.class_id = n.class_id
			AND ($3::text[] IS NULL OR cardinality($3::text[]) = 0 OR n.owner != ALL($3))
		WHERE 
			($5 = true OR n.owner != i.owner)
		GROUP BY creator, collector
	) AS r
	WHERE
		collector = ANY($1)
		AND rank <= $2
	ORDER BY rank
	;
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()

	rows, err := conn.Query(ctx, sql,
		collectorVariations, q.Top, ignoreListVariations, q.AllIscnVersions, q.IncludeOwner,
	)
	if err != nil {
		logger.L.Errorw("failed to query collector top ranked creators list", "error", err, "q", q)
		err = fmt.Errorf("failed to query collector top ranked creators list: %w", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var creator CollectorTopRankedCreator
		if err = rows.Scan(&creator.Creator, &creator.Rank); err != nil {
			return
		}
		res.Creators = append(res.Creators, creator)
	}
	return
}

func GetClassesOwners(conn *pgxpool.Conn, q QueryClassesOwnersRequest) (QueryClassesOwnersResponse, error) {
	ownersVariations := utils.ConvertAddressArrayPrefixes(q.Owners, AddressPrefixes)
	sql := `
		SELECT DISTINCT owner, class_id
		FROM nft
		WHERE class_id = ANY($1)
			AND ($2::text[] IS NULL OR cardinality($2::text[]) = 0 OR owner = ANY($2))
		ORDER BY owner, class_id
;
	`
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	rows, err := conn.Query(ctx, sql, q.ClassIds, ownersVariations)
	if err != nil {
		logger.L.Errorw("Failed to query nft classes owners", "error", err, "q", q)
		return QueryClassesOwnersResponse{}, fmt.Errorf("error on query nft classes owners: %w", err)
	}

	classIds := make(map[string][]string)
	for rows.Next() {
		var ownerRes string
		var classId string
		if err := rows.Scan(&ownerRes, &classId); err != nil {
			logger.L.Errorw("failed to scan owner address and class ID", "error", err)
			return QueryClassesOwnersResponse{}, fmt.Errorf("failed to scan owner address and class ID: %w", err)
		}
		convertedOwner, err := utils.ConvertAddressPrefix(ownerRes, MainAddressPrefix)
		if err != nil {
			logger.L.Errorw("failed to convert address prefix when processing QueryClassesOwnersRequest", "error", err, "owner", ownerRes, "class_id", classId, "request", q)
			// non-critical error, just skip this row
			continue
		}
		classIds[convertedOwner] = append(classIds[convertedOwner], classId)
	}
	return QueryClassesOwnersResponse{Owners: classIds}, nil
}
