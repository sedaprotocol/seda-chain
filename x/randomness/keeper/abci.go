package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/sedaprotocol/seda-chain/cmd/seda-chaind/utils"
	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

func PrepareProposalHandler(
	txConfig client.TxConfig,
	vrfSigner utils.VRFSigner,
	keeper Keeper,
	authKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		if vrfSigner == nil {
			return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, nil
		}

		// TO-DO run DefaultProposalHandler.PrepareProposalHandler first?

		// alpha = (seed_{i-1} || timestamp)
		prevSeed := keeper.GetSeed(ctx)
		if prevSeed == "" {
			panic("seed should never be empty")
		}
		timestamp, err := req.Time.MarshalBinary()
		if err != nil {
			return nil, err
		}
		alpha := append([]byte(prevSeed), timestamp...)

		// produce VRF proof
		pi, beta, err := vrfSigner.VRFProve(alpha)
		if err != nil {
			return nil, err
		}

		// create a NewSeed tx
		pubKey, err := keeper.GetValidatorVRFPubKey(ctx, sdk.ConsAddress(req.ProposerAddress).String())
		if err != nil {
			return nil, err
		}
		account := authKeeper.GetAccount(ctx, sdk.AccAddress(pubKey.Address().Bytes()))
		err = account.SetPubKey(pubKey) // checked later when signing tx with VRF key
		if err != nil {
			return nil, err
		}
		newSeedTx, err := generateAndSignNewSeedTx(ctx, txConfig, vrfSigner, account, &types.MsgNewSeed{
			Proposer: sdk.AccAddress(req.ProposerAddress).String(),
			Pi:       hex.EncodeToString(pi),
			Beta:     hex.EncodeToString(beta),
		})
		if err != nil {
			return nil, err
		}

		// prepend to list of txs and return
		res := new(abci.ResponsePrepareProposal)
		res.Txs = append([][]byte{newSeedTx}, req.Txs...)
		return res, nil
	}
}

func ProcessProposalHandler(
	txConfig client.TxConfig,
	vrfSigner utils.VRFSigner,
	keeper Keeper,
	stakingKeeper types.StakingKeeper,
) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		msg, err := decodeNewSeedTx(txConfig, req.Txs[0])
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		if msg.Proposer != string(sdk.AccAddress(req.ProposerAddress).String()) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT},
				fmt.Errorf("the NewSeed transaction must be from the block proposer")
		}

		// TO-DO run DefaultProposalHandler.ProcessProposalHandler first?
		// TO-DO Validate()?

		// get block proposer's validator public key
		pubKey, err := keeper.GetValidatorVRFPubKey(ctx, sdk.ConsAddress(req.ProposerAddress).String())
		if err != nil {
			return nil, err
		}

		prevSeed := keeper.GetSeed(ctx)
		if prevSeed == "" {
			panic("seed should never be empty")
		}
		timestamp, err := req.Time.MarshalBinary()
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}
		alpha := append([]byte(prevSeed), timestamp...)

		pi, err := hex.DecodeString(msg.Pi)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// verify VRF proof
		beta, err := vrfSigner.VRFVerify(pubKey.Bytes(), pi, alpha)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// sanity check
		msgBeta, err := hex.DecodeString(msg.Beta)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}
		if !bytes.Equal(beta, msgBeta) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// generateAndSignNewSeedTx generates and signs a transaction containing
// a given NewSeed message. It returns a transaction encoded into bytes.
func generateAndSignNewSeedTx(ctx sdk.Context, txConfig client.TxConfig, vrfSigner utils.VRFSigner, account sdk.AccountI, msg *types.MsgNewSeed) ([]byte, error) {
	// build a transaction containing the given message
	txBuilder := txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	txBuilder.SetGasLimit(200000) // TO-DO what number to put here?
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetFeePayer(account.GetAddress())

	// sign the transaction
	sig, err := vrfSigner.SignTransaction(
		ctx,
		txBuilder,
		txConfig,
		txsigning.SignMode_SIGN_MODE_DIRECT,
		account,
	)
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	tx := txBuilder.GetTx()
	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func decodeNewSeedTx(txConfig client.TxConfig, txBytes []byte) (*types.MsgNewSeed, error) {
	tx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return nil, err
	}
	msgNewSeed, ok := msgs[0].(*types.MsgNewSeed)
	if !ok {
		return nil, err
	}
	return msgNewSeed, nil
}
