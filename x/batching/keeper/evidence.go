package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"

	"cosmossdk.io/x/evidence/exported"
	evidencetypes "cosmossdk.io/x/evidence/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func NewBatchDoubleSignHandler(keeper Keeper) func(ctx context.Context, evidence exported.Evidence) error {
	return func(ctx context.Context, evidence exported.Evidence) error {
		return keeper.handleEvidence(ctx, evidence.(*types.BatchDoubleSign))
	}
}

func (k *Keeper) handleEvidence(ctx context.Context, evidence *types.BatchDoubleSign) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sdkCtx.Logger().Info("received batch double sign evidence", "batch number", evidence.BatchNumber, "operator address", evidence.OperatorAddress, "result root", evidence.DataResultRoot, "validator root", evidence.ValidatorRoot, "proving metadata hash", evidence.ProvingMetadataHash, "proving scheme index", evidence.ProvingSchemeIndex)

	// Validate that a batch exists for the given batch height.
	batch, err := k.GetBatchByBatchNumber(ctx, evidence.BatchNumber)
	if err != nil {
		return err
	}

	// Validate the signed batch is different from what was recorded on chain.
	fraudulentBatchID, err := evidence.GetBatchID()
	if err != nil {
		return err
	}

	if bytes.Equal(batch.BatchId, fraudulentBatchID) {
		return fmt.Errorf("batch IDs are the same")
	}

	// Currently we only support secp256k1 signatures so no need to check which proving scheme the validator used.
	signatureAddr, err := k.recoverEthAddressFromSecp256k1Signature(fraudulentBatchID, evidence.Signature)
	if err != nil {
		return err
	}

	// Retrieve the validator entry from the previous batch, as they might have changed their public key in the
	// fraudulent batch.
	validatorEthAddr, err := k.getEthAddressForBatch(ctx, evidence.BatchNumber-1, evidence.OperatorAddress)
	if err != nil {
		return err
	}

	// If the recovered address matches the validator entry they have committed a double sign.
	if !bytes.Equal(validatorEthAddr, signatureAddr) {
		return fmt.Errorf("recovered address does not match validator entry. Recorded: %s, Got: %s", hex.EncodeToString(validatorEthAddr), hex.EncodeToString(signatureAddr))
	}

	sdkCtx.Logger().Info("confirmed double batch sign", "validator", evidence.OperatorAddress, "batch number", evidence.BatchNumber, "block height", evidence.BlockHeight)

	// Reject evidence if the double-sign is too old. Evidence is considered stale
	// if the difference in number of blocks is greater than the allowed
	// parameter defined.
	// NOTE: The default max age is way larger than the default historical entries,
	// we should research if it's possible to use the CometBFT state to check the validator
	// power at the time of the double sign.
	infractionHeight := batch.BlockHeight
	ageBlocks := sdkCtx.BlockHeader().Height - infractionHeight
	cp := sdkCtx.ConsensusParams()
	if cp.Evidence != nil {
		if ageBlocks > cp.Evidence.MaxAgeNumBlocks {
			sdkCtx.Logger().Info(
				"ignored double batch sign; evidence too old",
				"validator", evidence.OperatorAddress,
				"infraction_height", infractionHeight,
				"max_age_num_blocks", cp.Evidence.MaxAgeNumBlocks,
			)
			return nil
		}
	}

	// We need to get the validator voting power at the time of the double sign.
	validator, err := k.getValidatorAtHeight(ctx, batch.BlockHeight, evidence.OperatorAddress)
	if err != nil {
		return err
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	if k.slashingKeeper.IsTombstoned(ctx, consAddr) {
		sdkCtx.Logger().Info("validator already tombstoned", "validator", evidence.OperatorAddress)
		return nil
	}

	slashFractionDoubleSign, err := k.slashingKeeper.SlashFractionDoubleSign(ctx)
	if err != nil {
		return err
	}

	// We use the staking keeper instead of the slashing keeper so we can emit the correct events.
	validatorPower := validator.GetConsensusPower(sdk.DefaultPowerReduction)
	coinsBurned, err := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		batch.BlockHeight,
		validatorPower,
		slashFractionDoubleSign,
	)
	if err != nil {
		return err
	}

	err = k.jailAndTombstone(ctx, consAddr, validator.IsJailed())
	if err != nil {
		return err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeOperatorAddress, evidence.OperatorAddress),
			sdk.NewAttribute(types.AttributePower, fmt.Sprintf("%d", validatorPower)),
			sdk.NewAttribute(types.AttributeReason, types.AttributeValueBatchDoubleSign),
			sdk.NewAttribute(types.AttributeBurnedCoins, coinsBurned.String()),
			sdk.NewAttribute(types.AttributeBatchNumber, fmt.Sprintf("%d", evidence.BatchNumber)),
			sdk.NewAttribute(types.AttributeProvingScheme, fmt.Sprintf("%d", evidence.ProvingSchemeIndex)),
		),
	)

	return nil
}

func (k *Keeper) recoverEthAddressFromSecp256k1Signature(batchID []byte, signature string) ([]byte, error) {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}

	signaturePubkey, err := crypto.Ecrecover(batchID, signatureBytes)
	if err != nil {
		return nil, err
	}

	signatureAddr, err := utils.PubKeyToEthAddress(signaturePubkey)
	if err != nil {
		return nil, err
	}

	return signatureAddr, nil
}

func (k *Keeper) getEthAddressForBatch(ctx context.Context, batchNumber uint64, operatorAddr string) ([]byte, error) {
	operatorAddrBytes, err := k.validatorAddressCodec.StringToBytes(operatorAddr)
	if err != nil {
		return nil, err
	}

	validatorEntry, err := k.GetValidatorTreeEntry(ctx, batchNumber, operatorAddrBytes)
	if err != nil {
		return nil, err
	}

	return validatorEntry.EthAddress, nil
}

// Retrieves the validator as it was at a given height, provided the height is within the historical info window.
func (k *Keeper) getValidatorAtHeight(ctx context.Context, height int64, operatorAddr string) (validator stakingtypes.Validator, err error) {
	historicalInfo, err := k.stakingKeeper.GetHistoricalInfo(ctx, height)
	if err != nil {
		return validator, err
	}

	for _, val := range historicalInfo.Valset {
		if val.OperatorAddress == operatorAddr {
			validator = val
			break
		}
	}

	if validator.OperatorAddress != operatorAddr {
		return validator, fmt.Errorf("validator not found")
	}

	return validator, nil
}

func (k *Keeper) jailAndTombstone(ctx context.Context, consAddr []byte, jailed bool) error {
	if !jailed {
		err := k.slashingKeeper.Jail(ctx, consAddr)
		if err != nil {
			return err
		}
	}

	err := k.slashingKeeper.JailUntil(ctx, consAddr, evidencetypes.DoubleSignJailEndTime)
	if err != nil {
		return err
	}

	err = k.slashingKeeper.Tombstone(ctx, consAddr)
	if err != nil {
		return err
	}

	return nil
}
