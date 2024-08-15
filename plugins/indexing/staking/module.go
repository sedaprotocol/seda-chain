package staking

import (
	"bytes"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const StoreKey = stakingtypes.StoreKey

type wrappedValidator struct {
	cdc       codec.Codec
	Validator stakingtypes.Validator
}

func (s wrappedValidator) MarshalJSON() ([]byte, error) {
	return s.cdc.MarshalJSON(&s.Validator)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, stakingtypes.ParamsKey); found {
		val, err := codec.CollValue[stakingtypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string              `json:"moduleName"`
			Params     stakingtypes.Params `json:"params"`
		}{
			ModuleName: "staking",
			Params:     val,
		}

		return types.NewMessage("module-params", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, stakingtypes.LastTotalPowerKey); found {
		power := sdk.IntProto{}
		err := cdc.Unmarshal(change.Value, &power)
		if err != nil {
			return nil, err
		}

		data := struct {
			Power string `json:"power"`
		}{
			Power: power.Int.String(),
		}

		return types.NewMessage("total-power", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, stakingtypes.LastValidatorPowerKey); found {
		_, key, err := sdk.ValAddressKey.Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := codec.CollValue[gogotypes.Int64Value](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ValidatorAddress string `json:"validatorAddress"`
			Power            int64  `json:"power"`
		}{
			ValidatorAddress: key.String(),
			Power:            val.Value,
		}

		return types.NewMessage("validator-power", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, stakingtypes.ValidatorsKey); found {
		val, err := codec.CollValue[stakingtypes.Validator](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("validator", &wrappedValidator{cdc: cdc, Validator: val}, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, stakingtypes.DelegationKey); found {
		if change.Delete {
			// The value of the delete change does not include the delegator nor the validator address.
			_, key, err := collections.PairKeyCodec(sdk.AccAddressKey, sdk.ValAddressKey).Decode(keyBytes)
			if err != nil {
				return nil, err
			}

			data := stakingtypes.Delegation{
				DelegatorAddress: key.K1().String(),
				ValidatorAddress: key.K2().String(),
				Shares:           math.LegacyZeroDec(),
			}

			return types.NewMessage("delegation", data, ctx), nil
		}

		val, err := codec.CollValue[stakingtypes.Delegation](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("delegation", val, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, stakingtypes.UnbondingDelegationKey); found {
		if change.Delete {
			// The value of the delete change does not include the delegator nor the validator address.
			_, key, err := collections.PairKeyCodec(sdk.AccAddressKey, sdk.ValAddressKey).Decode(keyBytes)
			if err != nil {
				return nil, err
			}

			data := stakingtypes.UnbondingDelegation{
				DelegatorAddress: key.K1().String(),
				ValidatorAddress: key.K2().String(),
				Entries:          nil,
			}

			return types.NewMessage("unbonding-delegation", data, ctx), nil
		}

		val, err := codec.CollValue[stakingtypes.UnbondingDelegation](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("unbonding-delegation", val, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, stakingtypes.RedelegationKey); found {
		if change.Delete {
			// The value of the delete change does not include the delegator nor the validator addresses.
			_, key, err := collections.TripleKeyCodec(sdk.AccAddressKey, sdk.ValAddressKey, sdk.ValAddressKey).Decode(keyBytes)
			if err != nil {
				return nil, err
			}

			data := stakingtypes.Redelegation{
				DelegatorAddress:    key.K1().String(),
				ValidatorSrcAddress: key.K2().String(),
				ValidatorDstAddress: key.K3().String(),
				Entries:             nil,
			}

			return types.NewMessage("redelegation", data, ctx), nil
		}

		val, err := codec.CollValue[stakingtypes.Redelegation](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("redelegation", val, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
