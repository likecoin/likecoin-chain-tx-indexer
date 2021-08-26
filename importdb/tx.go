package importdb

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/likecoin/likechain/app"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/types"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var encodingConfig = app.MakeEncodingConfig()

func newResponseResultTx(txHash bytes.HexBytes, res *abci.TxResult, tx *codecTypes.Any, timestamp string) (sdk.TxResponse, error) {
	parsedLogs, _ := sdk.ParseABCILogs(res.Result.Log)
	stdTx, err := convertToStdTx(tx.Value)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// Pack the amino stdTx into the TxResponse's Any.
	txAny := codecTypes.UnsafePackAny(stdTx)

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
		Tx:        txAny,
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

func convertToStdTx(txBytes []byte) (legacytx.StdTx, error) {
	txI, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return legacytx.StdTx{}, err
	}

	tx, ok := txI.(signing.Tx)
	if !ok {
		return legacytx.StdTx{}, fmt.Errorf("%+v is not backwards compatible with %T", tx, legacytx.StdTx{})
	}

	return clienttx.ConvertTxToStdTx(encodingConfig.Amino, tx)
}
