package base

import (
	"encoding/hex"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

type wrappedTx struct {
	cdc codec.Codec
	Tx  *tx.Tx
}

func (s wrappedTx) MarshalJSON() ([]byte, error) {
	return s.cdc.MarshalJSON(s.Tx)
}

func ExtractTransactionUpdates(ctx *types.BlockContext, cdc codec.Codec, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) ([]*types.Message, error) {
	messages := make([]*types.Message, 0, len(req.Txs))

	timestamp := req.Time

	for index := range req.Txs {
		txBytes := req.Txs[index]
		txResult := res.TxResults[index]
		txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(txBytes)))

		var tx tx.Tx
		if err := cdc.Unmarshal(txBytes, &tx); err != nil {
			return nil, err
		}

		signersBytes, _, err := tx.GetSigners(cdc)
		if err != nil {
			return nil, err
		}

		signers := make([]string, 0, len(signersBytes))
		for _, signerBytes := range signersBytes {
			var signer sdk.AccAddress
			if err := signer.Unmarshal(signerBytes); err != nil {
				return nil, err
			}
			signers = append(signers, signer.String())
		}

		data := struct {
			Hash    string             `json:"hash"`
			Time    time.Time          `json:"time"`
			Tx      *wrappedTx         `json:"tx"`
			Signers []string           `json:"signers"`
			Result  *abci.ExecTxResult `json:"result"`
		}{
			Hash:    txHash,
			Time:    timestamp,
			Tx:      &wrappedTx{cdc: cdc, Tx: &tx},
			Signers: signers,
			Result:  txResult,
		}

		messages = append(messages, types.NewMessage("tx", data, ctx))
	}

	return messages, nil
}
