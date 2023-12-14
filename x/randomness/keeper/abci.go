package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"

	vrf "github.com/sedaprotocol/vrf-go"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

func PrepareProposalHandler(
	txConfig client.TxConfig,
	keeper Keeper,
	authKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	mempool mempool.Mempool,
) sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
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

		// prepare secret key
		secretKey, err := readPrivKey("/Users/hykim/.seda-chain/config/priv_validator_key.json")
		if err != nil {
			return nil, err
		}

		// produce VRF proof
		k256vrf := vrf.NewK256VRF(0xFE)
		pi, err := k256vrf.Prove(secretKey.Bytes(), alpha)
		if err != nil {
			return nil, err
		}

		// debug
		fmt.Println(alpha)
		fmt.Println(secretKey.Bytes())
		fmt.Println(pi)

		beta, err := k256vrf.ProofToHash(pi)
		if err != nil {
			return nil, err
		}
		// // zero it out
		// for i := range secretKey {
		// 	secretKey[i] = 0
		// }

		validator, err := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(req.ProposerAddress))
		if err != nil {
			return nil, err
		}
		publicKey, err := validator.ConsPubKey()
		if err != nil {
			return nil, err
		}
		account := authKeeper.GetAccount(ctx, sdk.AccAddress(publicKey.Address().Bytes()))

		newSeedTx, _, err := encodeNewSeedTx(ctx, txConfig, secretKey, publicKey, account, &types.MsgNewSeed{
			Proposer: sdk.AccAddress(req.ProposerAddress).String(),
			Pi:       hex.EncodeToString(pi),
			Beta:     hex.EncodeToString(beta),
		})
		if err != nil {
			return nil, err
		}

		// TO-DO mempool
		// err = mempool.Insert(ctx, tx)
		// if err != nil {
		// 	return nil, err
		// }

		// prepend to list of txs and return
		res := new(abci.ResponsePrepareProposal)
		res.Txs = append([][]byte{newSeedTx}, req.Txs...)
		return res, nil
	}
}

func ProcessProposalHandler(
	txConfig client.TxConfig,
	keeper Keeper,
	stakingKeeper types.StakingKeeper,
) sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		msg, err := decodeNewSeedTx(txConfig, req.Txs[0])
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// TO-DO run DefaultProposalHandler.ProcessProposalHandler first?
		// TO-DO Validate()?
		// TO-DO Ensure that msg was signed by block proposer?

		// get block proposer's validator public key
		validator, err := stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(req.ProposerAddress))
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}
		publicKey, err := validator.ConsPubKey()
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
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
		k256vrf := vrf.NewK256VRF(0xFE)
		beta, err := k256vrf.Verify(publicKey.Bytes(), pi, alpha)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// sanity check
		msgBeta, err := hex.DecodeString(msg.Beta)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}
		if !bytes.Equal(beta, msgBeta) {
			panic(err)
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func encodeNewSeedTx(ctx sdk.Context, txConfig client.TxConfig, privKey crypto.PrivKey, pubKey cryptotypes.PubKey, account sdk.AccountI, msg *types.MsgNewSeed) ([]byte, sdk.Tx, error) {
	txBuilder := txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, nil, err
	}

	txBuilder.SetFeePayer(account.GetAddress())
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200000) // TO-DO what number to put here?

	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
		PubKey:        pubKey,
		Address:       account.GetAddress().String(),
	}

	// TO-DO re-examine signing logic

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sig := txsigning.SignatureV2{
		PubKey: pubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: account.GetSequence(),
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, nil, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		txsigning.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, nil, err
	}

	sigBytes, err := privKey.Sign(bytesToSign)
	// sigBytes, err := v.privateKey.Sign(bytesToSign)
	if err != nil {
		return nil, nil, err
	}

	sig = txsigning.SignatureV2{
		PubKey: pubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: account.GetSequence(),
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, nil, err
	}

	signedTx := txBuilder.GetTx()
	txBytes, err := txConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, nil, err
	}

	tx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, nil, err
	}
	return txBytes, tx, nil
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

func readPrivKey(keyFilePath string) (crypto.PrivKey, error) {
	keyJSONBytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, err
	}
	pvKey := privval.FilePVKey{}
	err = cmtjson.Unmarshal(keyJSONBytes, &pvKey)
	if err != nil {
		return nil, fmt.Errorf("error reading PrivValidator key from %v: %w", keyFilePath, err)
	}
	return pvKey.PrivKey, nil
}
