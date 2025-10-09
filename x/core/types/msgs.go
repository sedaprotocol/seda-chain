package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
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

	if m.ReplicationFactor == 0 {
		return ErrZeroReplicationFactor
	}
	// Ensure that the replication factor fits within 16 bits (unsigned).
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

// MsgHash returns the hex-encoded hash of the PostDataRequest message to be used
// as the data request ID.
func (m *MsgPostDataRequest) MsgHash() (string, error) {
	execProgramIDBytes, err := hex.DecodeString(m.ExecProgramID)
	if err != nil {
		return "", err
	}
	tallyProgramIDBytes, err := hex.DecodeString(m.TallyProgramID)
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
	//nolint:gosec // G115: Replication factor is guaranteed to fit in uint16.
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

func (m MsgStake) MsgHash(chainID string, sequenceNum uint64) ([]byte, error) {
	memoBytes, err := base64.StdEncoding.DecodeString(m.Memo)
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

	allBytes := append([]byte{}, []byte("stake")...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

func (m MsgUnstake) MsgHash(chainID string, sequenceNum uint64) ([]byte, error) {
	allBytes := append([]byte{}, []byte("unstake")...)
	allBytes = append(allBytes, []byte(chainID)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

func (m MsgWithdraw) MsgHash(chainID string, sequenceNum uint64) ([]byte, error) {
	allBytes := append([]byte{}, []byte("withdraw")...)
	allBytes = append(allBytes, []byte(m.WithdrawAddress)...)
	allBytes = append(allBytes, []byte(chainID)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

func (m MsgCommit) MsgHash(chainID string, drHeight int64) ([]byte, error) {
	drHeightBytes := make([]byte, 8)
	//nolint:gosec // G115: Block height is never negative.
	binary.BigEndian.PutUint64(drHeightBytes, uint64(drHeight))

	allBytes := append([]byte{}, []byte("commit_data_result")...)
	allBytes = append(allBytes, []byte(m.DrID)...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, m.Commit...)
	allBytes = append(allBytes, []byte(chainID)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

func (m MsgReveal) Validate(config DataRequestConfig, replicationFactor uint32) error {
	// Ensure that the exit code fits within 8 bits (unsigned).
	if m.RevealBody.ExitCode > uint32(^uint8(0)) {
		return ErrInvalidRevealExitCode
	}

	revealSizeLimit := config.DrRevealSizeLimitInBytes / replicationFactor
	if len(m.RevealBody.Reveal) > int(revealSizeLimit) {
		return ErrRevealTooBig.Wrapf("%d bytes > %d bytes", len(m.RevealBody.Reveal), revealSizeLimit)
	}

	for _, key := range m.RevealBody.ProxyPubKeys {
		_, err := hex.DecodeString(key)
		if err != nil {
			return ErrInvalidProxyPublicKey.Wrapf("%s", err.Error())
		}
	}
	return nil
}

func (m MsgReveal) MsgHash(chainID string) ([]byte, error) {
	revealBodyHash, err := m.RevealBody.RevealBodyHash()
	if err != nil {
		return nil, err
	}

	allBytes := append([]byte("reveal_data_result"), revealBodyHash...)
	allBytes = append(allBytes, []byte(chainID)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

// RevealHash computes a hash of reveal contents to be used as a commitment
// by the executors.
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
	revealHasher.Write(rb.Reveal)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	idBytes, err := hex.DecodeString(rb.DrID)
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

// Parts extracts the hex-encoded public key, drID, and proof from the query request.
func (q QueryGetExecutorEligibilityRequest) Parts() (string, string, string, error) {
	data, err := base64.StdEncoding.DecodeString(q.Data)
	if err != nil {
		return "", "", "", err
	}
	return string(data[:66]), string(data[67:131]), string(data[132:]), nil
}

func (q QueryGetExecutorEligibilityRequest) MsgHash(chainID string) []byte {
	_, drIDBytes, _, _ := q.Parts()
	allBytes := append([]byte{}, []byte("is_executor_eligible")...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, []byte(chainID)...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}
