package wasm

import (
	"encoding/json"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"

	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
)

const (
	QueryPathStakerAndSeq      = "/sedachain.core.v1.Query/StakerAndSeq"
	QueryPathStakingConfig     = "/sedachain.core.v1.Query/StakingConfig"
	QueryPathDataRequestConfig = "/sedachain.core.v1.Query/DataRequestConfig"
)

type CoreContractQuery struct {
	GetStakerAndSeq  *GetStakerAndSeq  `json:"get_staker_and_seq"`
	GetStakingConfig *GetStakingConfig `json:"get_staking_config"`
}

type GetStakerAndSeq struct {
	PublicKey string `json:"public_key"`
}

type GetStakerAndSeqResponse struct {
	Seq    string `json:"seq"`
	Staker struct {
		Memo                    []byte `json:"memo"`
		TokensPendingWithdrawal string `json:"tokens_pending_withdrawal"`
		TokensStaked            string `json:"tokens_staked"`
	} `json:"staker"`
}

func (g GetStakerAndSeq) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryStakerAndSeqRequest{
		PublicKey: g.PublicKey,
	}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathStakerAndSeq, nil
}

func (g GetStakerAndSeq) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryStakerAndSeqResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetStakerAndSeqResponse{
		Seq: strconv.FormatUint(res.SequenceNum, 10),
		Staker: struct {
			Memo                    []byte `json:"memo"`
			TokensPendingWithdrawal string `json:"tokens_pending_withdrawal"`
			TokensStaked            string `json:"tokens_staked"`
		}{
			Memo:                    []byte(res.Staker.Memo),
			TokensPendingWithdrawal: res.Staker.PendingWithdrawal.String(),
			TokensStaked:            res.Staker.Staked.String(),
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

type GetStakingConfig struct{}

type GetStakingConfigResponse struct {
	StakingConfig struct {
		MinimumStake     string `json:"minimum_stake"`
		AllowlistEnabled bool   `json:"allowlist_enabled"`
	} `json:"staking_config"`
}

func (g GetStakingConfig) ToModuleQuery() ([]byte, string, error) {
	query := &coretypes.QueryStakingConfigRequest{}
	queryProto, err := query.Marshal()
	if err != nil {
		return nil, "", err
	}
	return queryProto, QueryPathStakingConfig, nil
}

func (g GetStakingConfig) FromModuleQuery(cdc codec.Codec, result []byte) ([]byte, error) {
	var res coretypes.QueryStakingConfigResponse
	err := cdc.Unmarshal(result, &res)
	if err != nil {
		return nil, err
	}

	response := GetStakingConfigResponse{
		StakingConfig: struct {
			MinimumStake     string `json:"minimum_stake"`
			AllowlistEnabled bool   `json:"allowlist_enabled"`
		}{
			MinimumStake:     res.StakingConfig.MinimumStake.String(),
			AllowlistEnabled: res.StakingConfig.AllowlistEnabled,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}
