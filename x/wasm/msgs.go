package wasm

import (
	"encoding/base64"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
)

type PostRequestResponsePayload struct {
	DrID   string `json:"dr_id"`
	Height uint64 `json:"height"`
}

type CoreContractMsg struct {
	AddToAllowList   *AddToAllowListMsg   `json:"add_to_allowlist"`
	Stake            *StakeMsg            `json:"stake"`
	PostDataRequest  *PostDataRequestMsg  `json:"post_data_request"`
	CommitDataResult *CommitDataResultMsg `json:"commit_data_result"`
	RevealDataResult *RevealDataResultMsg `json:"reveal_data_result"`
}

type AddToAllowListMsg struct {
	PublicKey string `json:"public_key"`
}

func (m AddToAllowListMsg) EncodeToSdkMsg(sender string) (sdk.Msg, error) {
	return &coretypes.MsgAddToAllowlist{
		Sender:    sender, // TODO ensure security
		PublicKey: m.PublicKey,
	}, nil
}

type StakeMsg struct {
	PublicKey string `json:"public_key"`
	Proof     string `json:"proof"`
	Memo      string `json:"memo"`
}

func (m StakeMsg) EncodeToSdkMsg(sender string, stake sdk.Coin) (sdk.Msg, error) {
	return &coretypes.MsgStake{
		Sender:    sender, // TODO ensure security
		PublicKey: m.PublicKey,
		Proof:     m.Proof,
		Memo:      m.Memo,
		Stake:     stake,
	}, nil
}

type PostDataRequestMsg struct {
	PostedDR       PostedDR `json:"posted_dr"`
	SedaPayload    string   `json:"seda_payload"`
	PaybackAddress string   `json:"payback_address"`
}

type PostedDR struct {
	Version           string `json:"version"`
	ExecProgramID     string `json:"exec_program_id"`
	ExecInputs        string `json:"exec_inputs"`
	ExecGasLimit      uint64 `json:"exec_gas_limit"`
	TallyProgramID    string `json:"tally_program_id"`
	TallyInputs       string `json:"tally_inputs"`
	TallyGasLimit     uint64 `json:"tally_gas_limit"`
	ReplicationFactor uint16 `json:"replication_factor"`
	ConsensusFilter   string `json:"consensus_filter"`
	GasPrice          string `json:"gas_price"`
	Memo              string `json:"memo"`
}

func (m PostDataRequestMsg) EncodeToSdkMsg(sender string, funds sdk.Coin) (sdk.Msg, error) {
	execInputs, err := base64.StdEncoding.DecodeString(m.PostedDR.ExecInputs)
	if err != nil {
		return nil, err
	}
	tallyInputs, err := base64.StdEncoding.DecodeString(m.PostedDR.TallyInputs)
	if err != nil {
		return nil, err
	}
	consensusFilter, err := base64.StdEncoding.DecodeString(m.PostedDR.ConsensusFilter)
	if err != nil {
		return nil, err
	}
	memo, err := base64.StdEncoding.DecodeString(m.PostedDR.Memo)
	if err != nil {
		return nil, err
	}
	sedaPayload, err := base64.StdEncoding.DecodeString(m.SedaPayload)
	if err != nil {
		return nil, err
	}
	paybackAddress, err := base64.StdEncoding.DecodeString(m.PaybackAddress)
	if err != nil {
		return nil, err
	}

	gasPriceInt, ok := math.NewIntFromString(m.PostedDR.GasPrice)
	if !ok {
		return nil, fmt.Errorf("failed to convert gas price to big.Int")
	}

	return &coretypes.MsgPostDataRequest{
		Sender:            sender,
		Funds:             funds,
		Version:           m.PostedDR.Version,
		ExecProgramID:     m.PostedDR.ExecProgramID,
		ExecInputs:        execInputs,
		ExecGasLimit:      m.PostedDR.ExecGasLimit,
		TallyProgramID:    m.PostedDR.TallyProgramID,
		TallyInputs:       tallyInputs,
		TallyGasLimit:     m.PostedDR.TallyGasLimit,
		ReplicationFactor: uint32(m.PostedDR.ReplicationFactor),
		ConsensusFilter:   consensusFilter,
		GasPrice:          gasPriceInt,
		Memo:              memo,
		SEDAPayload:       sedaPayload,
		PaybackAddress:    paybackAddress,
	}, nil
}

type CommitDataResultMsg struct {
	DrID       string `json:"dr_id"`
	Commitment string `json:"commitment"`
	PublicKey  string `json:"public_key"`
	Proof      string `json:"proof"`
}

func (m CommitDataResultMsg) EncodeToSdkMsg(sender string) (sdk.Msg, error) {
	return &coretypes.MsgCommit{
		Sender:    sender, // TODO ensure security
		DrID:      m.DrID,
		Commit:    m.Commitment,
		PublicKey: m.PublicKey,
		Proof:     m.Proof,
	}, nil
}

type RevealDataResultMsg struct {
	RevealBody coretypes.RevealBody `json:"reveal_body"`
	PublicKey  string               `json:"public_key"`
	Proof      string               `json:"proof"`
	Stderr     []string             `json:"stderr"`
	Stdout     []string             `json:"stdout"`
}

func (m RevealDataResultMsg) EncodeToSdkMsg(sender string) (sdk.Msg, error) {
	return &coretypes.MsgReveal{
		Sender: sender, // TODO ensure security
		RevealBody: &coretypes.RevealBody{
			DrID:          m.RevealBody.DrID,
			DrBlockHeight: m.RevealBody.DrBlockHeight,
			ExitCode:      m.RevealBody.ExitCode,
			GasUsed:       m.RevealBody.GasUsed,
			Reveal:        m.RevealBody.Reveal,
			ProxyPubKeys:  m.RevealBody.ProxyPubKeys,
		},
		PublicKey: m.PublicKey,
		Proof:     m.Proof,
		Stderr:    m.Stderr,
		Stdout:    m.Stdout,
	}, nil
}
