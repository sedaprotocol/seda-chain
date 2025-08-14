package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"

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

// Validate validates the PostDataRequest message based on the given data
// request configurations.
func (m MsgPostDataRequest) Validate(config DataRequestConfig) error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("%s", err.Error())
	}

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

	if _, err := hex.DecodeString(m.ExecProgramId); err != nil {
		return ErrInvalidExecProgramID
	}
	if _, err := hex.DecodeString(m.TallyProgramId); err != nil {
		return ErrInvalidTallyProgramID
	}
	if len(m.ExecProgramId) != 64 {
		return ErrInvalidLengthExecProgramID.Wrapf("given ID is %d characters long", len(m.ExecProgramId))
	}
	if len(m.TallyProgramId) != 64 {
		return ErrInvalidLengthTallyProgramID.Wrapf("given ID is %d characters long", len(m.TallyProgramId))
	}

	// TODO
	// // Ensure the version only consists of Major.Minor.Patch
	// if !self.posted_dr.version.pre.is_empty() || !self.posted_dr.version.build.is_empty() {
	// 	return Err(ContractError::DataRequestVersionInvalid);
	// }

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

// MsgHash returns the hex-encoded hash of the PostDataRequest message to be used
// as the data request ID.
func (m *MsgPostDataRequest) MsgHash() (string, error) {
	execProgramIDBytes, err := hex.DecodeString(m.ExecProgramId)
	if err != nil {
		return "", err
	}
	tallyProgramIDBytes, err := hex.DecodeString(m.TallyProgramId)
	if err != nil {
		return "", err
	}

	execInputsHasher := sha3.NewLegacyKeccak256()
	execInputsHasher.Write(m.ExecInputs)
	execInputsHash := execInputsHasher.Sum(nil)

	tallyInputsHasher := sha3.NewLegacyKeccak256()
	tallyInputsHasher.Write(m.TallyInputs)
	tallyInputsHash := tallyInputsHasher.Sum(nil)

	consensusFilterHasher := sha3.NewLegacyKeccak256()
	consensusFilterHasher.Write(m.ConsensusFilter)
	consensusFilterHash := consensusFilterHasher.Sum(nil)

	memoHasher := sha3.NewLegacyKeccak256()
	memoHasher.Write(m.Memo)
	memoHash := memoHasher.Sum(nil)

	execGasLimitBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(execGasLimitBytes, m.ExecGasLimit)
	tallyGasLimitBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tallyGasLimitBytes, m.TallyGasLimit)
	replicationFactorBytes := make([]byte, 2)
	//nolint:gosec // G115: No overflow guaranteed by validation logic.
	binary.BigEndian.PutUint16(replicationFactorBytes, uint16(m.ReplicationFactor))

	dataRequestHasher := sha3.NewLegacyKeccak256()
	dataRequestHasher.Write([]byte(m.Version))
	dataRequestHasher.Write(execProgramIDBytes)
	dataRequestHasher.Write(execInputsHash)
	dataRequestHasher.Write(execGasLimitBytes)
	dataRequestHasher.Write(tallyProgramIDBytes)
	dataRequestHasher.Write(tallyInputsHash)
	dataRequestHasher.Write(tallyGasLimitBytes)
	dataRequestHasher.Write(replicationFactorBytes)
	dataRequestHasher.Write(consensusFilterHash)
	dataRequestHasher.Write(m.GasPrice.BigInt().Bytes())
	dataRequestHasher.Write(memoHash)

	return hex.EncodeToString(dataRequestHasher.Sum(nil)), nil
}

// TODO Remove contractAddr
// StakeHash computes the stake hash.
func (m MsgStake) MsgHash(_, chainID string, sequenceNum uint64) ([]byte, error) {
	memoBytes, err := hex.DecodeString(m.Memo)
	if err != nil {
		return nil, err
	}
	memoHasher := sha3.NewLegacyKeccak256()
	memoHasher.Write(memoBytes)
	memoHash := memoHasher.Sum(nil)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)

	allBytes := append([]byte{}, "stake"...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, chainID...)
	// allBytes = append(allBytes, contractAddr...) // TODO Do not include contractAddr
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

// TODO Remove contractAddr
// CommitHash computes the commit hash.
func (m MsgCommit) MsgHash(_, chainID string, drHeight int64) ([]byte, error) {
	drHeightBytes := make([]byte, 8)
	//nolint:gosec // G115: Block height is never negative.
	binary.BigEndian.PutUint64(drHeightBytes, uint64(drHeight))

	allBytes := append([]byte{}, "commit_data_result"...)
	allBytes = append(allBytes, m.DrId...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, m.Commit...)
	allBytes = append(allBytes, chainID...)
	// allBytes = append(allBytes, contractAddr...) // TODO Do not include contractAddr

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

// TODO Remove contractAddr
func (m MsgReveal) MsgHash(_, chainID string) ([]byte, error) {
	revealBodyHash, err := m.RevealBody.RevealBodyHash()
	if err != nil {
		return nil, err
	}

	allBytes := append([]byte("reveal_data_result"), revealBodyHash...)
	allBytes = append(allBytes, chainID...)
	// allBytes = append(allBytes, contractAddr...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

// RevealHash computes the hash of the reveal contents. This hash is used by
// executors as their commit value.
func (m MsgReveal) RevealHash() ([]byte, error) {
	revealBodyHash, err := m.RevealBody.RevealBodyHash()
	if err != nil {
		return nil, err
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("reveal_message"))
	hasher.Write(revealBodyHash)
	hasher.Write([]byte(m.PublicKey))
	hasher.Write([]byte(m.Proof))
	hasher.Write([]byte(strings.Join(m.Stderr, "")))
	hasher.Write([]byte(strings.Join(m.Stdout, "")))

	return hasher.Sum(nil), nil
}

func (rb RevealBody) RevealBodyHash() ([]byte, error) {
	revealHasher := sha3.NewLegacyKeccak256()
	revealBytes, err := base64.StdEncoding.DecodeString(rb.Reveal)
	if err != nil {
		return nil, err
	}
	revealHasher.Write(revealBytes)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	idBytes, err := hex.DecodeString(rb.DrId)
	if err != nil {
		return nil, err
	}
	hasher.Write(idBytes)

	reqHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reqHeightBytes, rb.DrBlockHeight)
	hasher.Write(reqHeightBytes)

	// TODO RevealBody validator should bind rb.ExitCode value?
	hasher.Write([]byte{byte(rb.ExitCode)})

	gasUsedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasUsedBytes, rb.GasUsed)
	hasher.Write(gasUsedBytes)

	hasher.Write(revealHash)

	proxyPubKeyHasher := sha3.NewLegacyKeccak256()
	for _, key := range rb.ProxyPubKeys {
		keyHasher := sha3.NewLegacyKeccak256()
		keyHasher.Write([]byte(key))
		proxyPubKeyHasher.Write(keyHasher.Sum(nil))
	}
	hasher.Write(proxyPubKeyHasher.Sum(nil))

	return hasher.Sum(nil), nil
}
