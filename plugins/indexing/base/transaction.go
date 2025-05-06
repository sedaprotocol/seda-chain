package base

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

type wrappedTx struct {
	cdc codec.Codec
	Tx  *tx.Tx
}

func (s wrappedTx) MarshalJSON() ([]byte, error) {
	return s.cdc.MarshalJSON(s.Tx)
}

func ExtractTransactionUpdates(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) ([]*types.Message, error) {
	messages := make([]*types.Message, 0, len(req.Txs))

	timestamp := req.Time

	logger.Debug(fmt.Sprintf("Number of transactions: %d", len(req.Txs)))
	for index := range req.Txs {
		txBytes := req.Txs[index]
		txResult := res.TxResults[index]
		txHash := strings.ToUpper(hex.EncodeToString(tmhash.Sum(txBytes)))
		logger.Trace(fmt.Sprintf("Processing transaction [%d] hash: %s", index, txHash))

		var tx tx.Tx
		// For some reason we sometimes get extended commit info that does serialise to a valid tx
		// so we need the additional check on the body to ensure the tx is valid. :(
		if err := cdc.Unmarshal(txBytes, &tx); err != nil || tx.Body == nil {
			logger.Trace("Error unmarshalling transaction, checking if it's an extended commit info")
			var extendedVotes abci.ExtendedCommitInfo
			if err := json.Unmarshal(txBytes, &extendedVotes); err != nil {
				logger.Trace("Error unmarshalling extended commit info")
				return nil, err
			}
			logger.Trace("Skipping extended votes bytes")
			continue
		}

		logger.Trace("Getting signers")
		signersBytes, _, err := tx.GetSigners(cdc)
		if err != nil {
			return nil, err
		}

		logger.Trace("Creating signers")
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
		logger.Trace(fmt.Sprintf("Transaction [%d] processed", index))
	}

	logger.Debug(fmt.Sprintf("Processed %d transactions", len(messages)))
	return messages, nil
}
