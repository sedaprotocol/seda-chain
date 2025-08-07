package types

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"

	"cosmossdk.io/math"
	"golang.org/x/crypto/sha3"
)

const (
	MinExecGasLimit      uint64 = 10_000_000_000_000 // 10 teraGas
	MinTallyGasLimit     uint64 = 10_000_000_000_000 // 10 teraGas
	MaxReplicationFactor int    = 100
)

var MinGasPrice = math.NewInt(2_000)

func isBigIntUint128(x *big.Int) bool {
	return x.Sign() >= 0 && x.BitLen() <= 128
}

// Validate validates the PostDataRequest message based on the given data
// request configurations.
func (m MsgPostDataRequest) Validate(config DataRequestConfig) error {
	if m.ReplicationFactor == 0 {
		return ErrZeroReplicationFactor
	}

	if !isBigIntUint128(m.GasPrice.BigInt()) {
		return ErrGasPriceTooHigh
	}
	if m.GasPrice.LT(MinGasPrice) {
		return ErrGasPriceTooLow.Wrapf("%s < %s", m.GasPrice, MinGasPrice)
	}
	if m.ExecGasLimit < MinExecGasLimit {
		return ErrExecGasLimitTooLow.Wrapf("%s < %s", m.ExecGasLimit, MinExecGasLimit)
	}
	if m.TallyGasLimit < MinTallyGasLimit {
		return ErrTallyGasLimitTooLow.Wrapf("%s < %s", m.TallyGasLimit, MinTallyGasLimit)
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
	if len(m.SedaPayload) > int(config.SedaPayloadLimitInBytes) {
		return ErrSedaPayloadLimitExceeded.Wrapf("%d bytes > %d bytes", len(m.SedaPayload), config.SedaPayloadLimitInBytes)
	}

	return nil
}

// Hash returns the hex-encoded hash of the PostDataRequest message to be used
// as the data request ID.
func (m *MsgPostDataRequest) Hash() (string, error) {
	execProgramIdBytes, err := hex.DecodeString(m.ExecProgramId)
	if err != nil {
		return "", err
	}
	tallyProgramIdBytes, err := hex.DecodeString(m.TallyProgramId)
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
	binary.BigEndian.PutUint16(replicationFactorBytes, uint16(m.ReplicationFactor))

	dataRequestHasher := sha3.NewLegacyKeccak256()
	dataRequestHasher.Write([]byte(m.Version))
	dataRequestHasher.Write(execProgramIdBytes)
	dataRequestHasher.Write(execInputsHash)
	dataRequestHasher.Write(execGasLimitBytes)
	dataRequestHasher.Write(tallyProgramIdBytes)
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
func (m MsgStake) StakeHash(contractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
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

	allBytes := append([]byte{}, []byte("stake")...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(contractAddr)...)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}

// TODO Remove contractAddr
// CommitHash computes the commit hash.
func (m MsgCommit) CommitHash(contractAddr, chainID string, drHeight uint64) ([]byte, error) {
	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	allBytes := append([]byte{}, []byte("commit_data_result")...)
	allBytes = append(allBytes, []byte(m.DrId)...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, m.Commitment...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(contractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}
