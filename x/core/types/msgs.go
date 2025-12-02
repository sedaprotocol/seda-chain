package types

import (
	"encoding/base64"
	"encoding/hex"
	"math/big"

	"golang.org/x/mod/semver"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MinExecGasLimit      uint64 = 10_000_000_000_000 // 10 teraGas
	MinTallyGasLimit     uint64 = 10_000_000_000_000 // 10 teraGas
	MaxReplicationFactor uint32 = 100
)

var MinGasPrice = math.NewInt(2_000)

func isBigIntUint128(x *big.Int) bool {
	return x.Sign() >= 0 && x.BitLen() <= 128
}

func (m MsgPostDataRequest) Validate(config DataRequestConfig) error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("%s", err.Error())
	}

	// Ensure that the replication factor is non-zero and fits within 16 bits (unsigned).
	if m.ReplicationFactor == 0 {
		return ErrZeroReplicationFactor
	}
	if m.ReplicationFactor > uint32(^uint16(0)) {
		return ErrReplicationFactorNotUint16
	}

	if m.GasPrice.IsNegative() || m.GasPrice.LT(MinGasPrice) {
		return ErrGasPriceTooLow.Wrapf("%s < %s", m.GasPrice, MinGasPrice)
	}
	if !isBigIntUint128(m.GasPrice.BigInt()) {
		return ErrGasPriceTooHigh
	}

	if m.ExecGasLimit < MinExecGasLimit {
		return ErrExecGasLimitTooLow.Wrapf("%d < %d", m.ExecGasLimit, MinExecGasLimit)
	}
	if m.TallyGasLimit < MinTallyGasLimit {
		return ErrTallyGasLimitTooLow.Wrapf("%d < %d", m.TallyGasLimit, MinTallyGasLimit)
	}

	if _, err := hex.DecodeString(m.ExecProgramID); err != nil {
		return ErrInvalidExecProgramID
	}
	if _, err := hex.DecodeString(m.TallyProgramID); err != nil {
		return ErrInvalidTallyProgramID
	}
	if len(m.ExecProgramID) != 64 {
		return ErrInvalidLengthExecProgramID.Wrapf("given ID is %d characters long", len(m.ExecProgramID))
	}
	if len(m.TallyProgramID) != 64 {
		return ErrInvalidLengthTallyProgramID.Wrapf("given ID is %d characters long", len(m.TallyProgramID))
	}

	// Ensure the version only consists of Major.Minor.Patch
	// with no prerelease or build suffixes.
	if !semver.IsValid("v"+m.Version) || semver.Prerelease("v"+m.Version) != "" || semver.Build("v"+m.Version) != "" {
		return ErrInvalidVersion.Wrapf("%s", m.Version)
	}

	if len(m.ExecInputs) > int(config.ExecInputLimitInBytes) {
		return ErrExecInputLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.ExecInputs), config.ExecInputLimitInBytes)
	}
	if len(m.TallyInputs) > int(config.TallyInputLimitInBytes) {
		return ErrTallyInputLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.TallyInputs), config.TallyInputLimitInBytes)
	}
	if len(m.ConsensusFilter) > int(config.ConsensusFilterLimitInBytes) {
		return ErrConsensusFilterLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.ConsensusFilter), config.ConsensusFilterLimitInBytes)
	}
	if len(m.Memo) > int(config.MemoLimitInBytes) {
		return ErrMemoLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.Memo), config.MemoLimitInBytes)
	}
	if len(m.PaybackAddress) > int(config.PaybackAddressLimitInBytes) {
		return ErrPaybackAddressLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.PaybackAddress), config.PaybackAddressLimitInBytes)
	}
	if len(m.SEDAPayload) > int(config.SEDAPayloadLimitInBytes) {
		return ErrSEDAPayloadLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.SEDAPayload), config.SEDAPayloadLimitInBytes)
	}

	return nil
}

func (m MsgStake) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", m.Sender)
	}
	_, err = base64.StdEncoding.DecodeString(m.Memo)
	if err != nil {
		return err
	}
	return nil
}

func (m MsgLegacyStake) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", m.Sender)
	}
	_, err = base64.StdEncoding.DecodeString(m.Memo)
	if err != nil {
		return err
	}
	return nil
}

func (m MsgReveal) Validate(config DataRequestConfig, replicationFactor uint32) error {
	// Ensure that the exit code fits within 8 bits (unsigned).
	if m.RevealBody.ExitCode > uint32(^uint8(0)) {
		return ErrInvalidRevealExitCode
	}

	_, err := hex.DecodeString(m.RevealBody.DrID)
	if err != nil {
		return ErrInvalidDataRequestIDHex.Wrapf("%s", err.Error())
	}

	revealSizeLimit := config.DrRevealSizeLimitInBytes / replicationFactor
	if len(m.RevealBody.Reveal) > int(revealSizeLimit) {
		return ErrRevealTooBig.Wrapf("%d bytes > %d bytes", len(m.RevealBody.Reveal), revealSizeLimit)
	}

	for _, key := range m.RevealBody.ProxyPubKeys {
		if key == "" {
			return ErrInvalidDataProxyPublicKey.Wrapf("public key is empty")
		}
		_, err := hex.DecodeString(key)
		if err != nil {
			return ErrInvalidDataProxyPublicKey.Wrapf("%s: %s", err.Error(), m.PublicKey)
		}
	}
	return nil
}

func (m MsgLegacyReveal) Validate(config DataRequestConfig, replicationFactor uint32) error {
	// Ensure that the exit code fits within 8 bits (unsigned).
	if m.RevealBody.ExitCode > uint32(^uint8(0)) {
		return ErrInvalidRevealExitCode
	}

	revealSizeLimit := config.DrRevealSizeLimitInBytes / replicationFactor
	if len(m.RevealBody.Reveal) > int(revealSizeLimit) {
		return ErrRevealTooBig.Wrapf("%d bytes > %d bytes", len(m.RevealBody.Reveal), revealSizeLimit)
	}

	for _, key := range m.RevealBody.ProxyPubKeys {
		if key == "" {
			return ErrInvalidDataProxyPublicKey.Wrapf("public key is empty")
		}
		_, err := hex.DecodeString(key)
		if err != nil {
			return ErrInvalidDataProxyPublicKey.Wrapf("%s: %s", err.Error(), key)
		}
	}
	return nil
}

func (m MsgAddToAllowlist) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", m.Sender)
	}
	if m.PublicKey == "" {
		return ErrInvalidStakerPublicKey.Wrapf("public key is empty")
	}
	_, err = hex.DecodeString(m.PublicKey)
	if err != nil {
		return ErrInvalidStakerPublicKey.Wrapf("%s: %s", err.Error(), m.PublicKey)
	}
	return nil
}

func (m MsgTransferOwnership) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", m.Sender)
	}
	_, err = sdk.AccAddressFromBech32(m.NewOwner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", m.NewOwner)
	}
	return nil
}

// Parts extracts the hex-encoded public key, drID, and proof from the query request.
//
// 0          66:67            130:131       195 byte
// | public_key : data_request_id :   proof   |
func (q QueryExecutorEligibilityRequest) Parts() (string, string, string, error) {
	if len(q.Data) != 194 {
		return "", "", "", ErrInvalidEligibilityProofLength.Wrapf("expected: %d, got: %d", 194, len(q.Data))
	}
	data, err := base64.StdEncoding.DecodeString(q.Data)
	if err != nil {
		return "", "", "", err
	}
	return string(data[:66]), string(data[67:131]), string(data[132:]), nil
}

// Parts for QueryLegacyExecutorEligibilityRequest is identical to that for
// QueryExecutorEligibilityRequest.
func (q QueryLegacyExecutorEligibilityRequest) Parts() (string, string, string, error) {
	if len(q.Data) != 194 {
		return "", "", "", ErrInvalidEligibilityProofLength.Wrapf("expected: %d, got: %d", 194, len(q.Data))
	}
	data, err := base64.StdEncoding.DecodeString(q.Data)
	if err != nil {
		return "", "", "", err
	}
	return string(data[:66]), string(data[66:130]), string(data[130:]), nil
}
