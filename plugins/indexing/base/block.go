package base

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

func ExtractBlockUpdate(ctx *types.BlockContext, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) (*types.Message, error) {
	hash := strings.ToUpper(hex.EncodeToString(req.Hash))
	txCount := len(req.Txs)
	if txCount > 0 {
		// Don't count extended votes as transactions
		var extendedVotes abci.ExtendedCommitInfo
		if err := json.Unmarshal(req.Txs[0], &extendedVotes); err == nil {
			txCount--
		}
	}
	proposerAddress, err := sdk.ConsAddressFromHex(hex.EncodeToString(req.ProposerAddress))
	if err != nil {
		return nil, err
	}

	var filteredEvents []abci.Event
	for _, event := range res.Events {
		skip := false
		for _, attribute := range event.Attributes {
			if attribute.Key == "amount" && attribute.Value == "" {
				skip = true
				break
			}
		}
		if !skip {
			filteredEvents = append(filteredEvents, event)
		}
	}

	data := struct {
		Hash            string       `json:"hash"`
		Time            time.Time    `json:"time"`
		TxCount         int          `json:"txCount"`
		ProposerAddress string       `json:"proposerAddress"`
		Events          []abci.Event `json:"events"`
	}{
		Hash:            hash,
		Time:            req.Time,
		TxCount:         txCount,
		ProposerAddress: proposerAddress.String(),
		Events:          filteredEvents,
	}

	return types.NewMessage("block", data, ctx), nil
}
