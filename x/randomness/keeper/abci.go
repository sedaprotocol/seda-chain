package keeper

import (
	"encoding/hex"
	"fmt"

	vrf "github.com/sedaprotocol/vrf-go"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

func (k *Keeper) EndBlocker(ctx sdk.Context) []abci.ValidatorUpdate {
	// defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	// k.SetSeed(ctx, strings.ToUpper(hex.EncodeToString(ctx.BlockHeader().AppHash)))
	return nil
}

func PrepareProposalHandler(
	txConfig client.TxConfig,
	keeper Keeper,
) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// Compute next seed using
		// - previous seed
		// - timestamp (or block height)
		// - private key
		prevSeed := keeper.GetSeed(ctx)
		timestamp, err := req.Time.MarshalBinary()
		if err != nil {
			return nil, err
		}
		alpha := append([]byte(prevSeed), timestamp...)

		// VRF call
		k256vrf := vrf.NewK256VRF(0xFE)
		secretKey, err := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140")
		if err != nil {
			return nil, err
		}

		pi, err := k256vrf.Prove(secretKey, alpha)
		if err != nil {
			return nil, err
		}

		proofStr := hex.EncodeToString(pi)
		fmt.Printf("\n\nsetting seed %s\n\n", proofStr)
		keeper.SetSeed(ctx, proofStr) // or add as transaction to be verified by validators before being stored.

		msgNewSeed := &types.MsgNewSeed{
			Seed:     "", // TO-DO technically unnecessary?
			Pi:       proofStr,
			Proposer: sdk.AccAddress(req.ProposerAddress).String(),
		}

		tx, err := EncodeMsgsIntoTxBytes(txConfig, msgNewSeed)
		if err != nil {
			return nil, err
		}

		res := new(abci.ResponsePrepareProposal)
		res.Txs = append(req.Txs, tx)

		// Q: Always add to the beginning so Process is more efficient??

		return res, nil
	}
}

func ProcessProposalHandler(
	txConfig client.TxConfig,
	keeper Keeper,
) sdk.ProcessProposalHandler {
	// // If the mempool is nil or NoOp we simply return ACCEPT,
	// // because PrepareProposal may have included txs that could fail verification.
	// _, isNoOp := h.mempool.(mempool.NoOpMempool)
	// if h.mempool == nil || isNoOp {
	// 	return NoOpProcessProposal()
	// }

	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// TO-DO replace loop with position-based check?
		for _, txBytes := range req.Txs {
			msg, err := DecodeNewSeedTx(txConfig.TxDecoder(), txBytes)
			if err != nil {
				continue
			}
			// if err != nil {
			// 	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			// }

			// TO-DO Validate()?

			// TO-DO Ensure that msg was signed by block proposer

			// TO-DO Get block proposer's validator public key
			proposer := sdk.AccAddress(req.ProposerAddress) // ValAddress or ConsAddress?
			fmt.Println(proposer.String())
			fmt.Println(msg.Proposer)

			/*
				prevSeed := keeper.GetSeed(ctx)
				timestamp, err := req.Time.MarshalBinary()
				if err != nil {
					return nil, err
				}
				alpha := append([]byte(prevSeed), timestamp...)

				// VRF call
				k256vrf := vrf.NewK256VRF(0xFE)

				pi, err := hex.DecodeString(msg.Pi)
				if err != nil {
					return nil, err
				}

				beta, err := k256vrf.Verify(publicKey, pi, alpha)
				if err != nil {
					return nil, err
				}
				betaStr := hex.EncodeToString(beta)  // computed

				if betaStr != msg.Seed {
					panic("no!")
				}

				keeper.SetSeed(ctx, msg.Seed)
			*/
			// TO-DO exit loop
		}

		// TO-DO what if new seed tx not found?

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func DecodeNewSeedTx(decoder sdk.TxDecoder, txBytes []byte) (*types.MsgNewSeed, error) {
	// Decode.
	tx, err := decoder(txBytes)
	if err != nil {
		return nil, err
	}

	// Check msg length.
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return nil, err
	}

	// Check msg type.
	msgNewSeed, ok := msgs[0].(*types.MsgNewSeed)
	if !ok {
		return nil, err
	}
	return msgNewSeed, nil
}

// EncodeMsgsIntoTxBytes encodes the given msgs into a single transaction.
func EncodeMsgsIntoTxBytes(txConfig client.TxConfig, msgs ...sdk.Msg) ([]byte, error) {
	txBuilder := txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}
