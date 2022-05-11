package importdb

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likecoin-chain/v2/app"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"

	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var encodingConfig = app.MakeEncodingConfig()

func newResponseResultTx(txHash bytes.HexBytes, res *abci.TxResult, tx *codecTypes.Any, timestamp string) (sdk.TxResponse, error) {
	parsedLogs, _ := sdk.ParseABCILogs(res.Result.Log)

	return sdk.TxResponse{
		TxHash:    txHash.String(),
		Height:    res.Height,
		Code:      res.Result.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.Result.Data)),
		RawLog:    res.Result.Log,
		Logs:      parsedLogs,
		Info:      res.Result.Info,
		GasWanted: res.Result.GasWanted,
		GasUsed:   res.Result.GasUsed,
		Tx:        tx,
		Timestamp: timestamp,
	}, nil
}

type intoAny interface {
	AsAny() *codecTypes.Any
}

func parseTx(txBytes []byte) (*codecTypes.Any, error) {
	txb, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	any := p.AsAny()
	return any, nil
}

func formatTxResult(txHash bytes.HexBytes, resTx *abci.TxResult, block *types.Block) (sdk.TxResponse, error) {
	tx, err := parseTx(resTx.Tx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return newResponseResultTx(txHash, resTx, tx, block.Header.Time.Format(time.RFC3339))
}
