package base

import (
	"encoding/hex"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/codec"
	txtype "github.com/cosmos/cosmos-sdk/types/tx"

	types "github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

type wrappedTx struct {
	cdc codec.Codec
	Tx  *txtype.Tx
}

func (s wrappedTx) MarshalJSON() ([]byte, error) {
	return s.cdc.MarshalJSON(s.Tx)
}

func ExtractTransactionUpdates(cdc codec.Codec, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) ([]*types.Message, error) {
	messages := make([]*types.Message, 0, len(req.Txs))

	timestamp := req.Time

	for index := range req.Txs {
		txBytes := req.Txs[index]
		txResult := res.TxResults[index]
		txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(txBytes)))

		var tx txtype.Tx
		if err := cdc.Unmarshal(txBytes, &tx); err != nil {
			return nil, err
		}

		data := struct {
			Hash   string             `json:"hash"`
			Time   time.Time          `json:"time"`
			Tx     *wrappedTx         `json:"tx"`
			Result *abci.ExecTxResult `json:"result"`
		}{
			Hash:   txHash,
			Time:   timestamp,
			Tx:     &wrappedTx{cdc: cdc, Tx: &tx},
			Result: txResult,
		}

		messages = append(messages, types.NewMessage("tx", data))
	}

	return messages, nil
}
