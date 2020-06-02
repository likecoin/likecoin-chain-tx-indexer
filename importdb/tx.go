package importdb

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/types"
)

var cdc *codec.Codec

func newResponseResultTx(txHash cmn.HexBytes, res *types.TxResult, tx sdk.Tx, timestamp string) sdk.TxResponse {
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
		Events:    sdk.StringifyEvents(res.Result.Events),
		Tx:        tx,
		Timestamp: timestamp,
	}
}

func parseTx(txBytes []byte) (sdk.Tx, error) {
	var tx authTypes.StdTx
	cdc.MustUnmarshalBinaryLengthPrefixed(txBytes, &tx)
	return tx, nil
}

func formatTxResult(txHash cmn.HexBytes, resTx *types.TxResult, block *types.Block) (sdk.TxResponse, error) {
	tx, err := parseTx(resTx.Tx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return newResponseResultTx(txHash, resTx, tx, block.Header.Time.Format(time.RFC3339)), nil
}
