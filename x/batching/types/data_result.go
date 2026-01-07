package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"
)

// TryHash returns a hex-encoded hash of the DataResult.
func (dr *DataResult) TryHash() (string, error) {
	hasher := sha3.NewLegacyKeccak256()

	versionBytes := []byte(dr.Version)
	hasher.Write(versionBytes)
	versionHash := hasher.Sum(nil)

	drIDBytes, err := hex.DecodeString(dr.DrId)
	if err != nil {
		return "", err
	}

	consensusByte := byte(0x00)
	if dr.Consensus {
		consensusByte = byte(0x01)
	}

	blockHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockHeightBytes, dr.BlockHeight)
	blockTimestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(blockTimestampBytes, dr.BlockTimestamp)

	exitCodeByte := byte(dr.ExitCode)

	hasher.Reset()
	hasher.Write(dr.Result)
	resultHash := hasher.Sum(nil)

	gasUsedBytes := make([]byte, 16)
	if dr.GasUsed != nil && !dr.GasUsed.IsNil() {
		gasUsedBigInt := dr.GasUsed.BigInt().Bytes()
		if len(gasUsedBigInt) > 16 {
			return "", errors.New("math.Int too large for u128")
		}
		copy(gasUsedBytes[16-len(gasUsedBigInt):], gasUsedBigInt)
	}

	paybackAddrBytes, err := base64.StdEncoding.DecodeString(dr.PaybackAddress)
	if err != nil {
		return "", err
	}
	hasher.Reset()
	hasher.Write(paybackAddrBytes)
	paybackAddrHash := hasher.Sum(nil)

	payloadBytes, err := base64.StdEncoding.DecodeString(dr.SedaPayload)
	if err != nil {
		return "", err
	}
	hasher.Reset()
	hasher.Write(payloadBytes)
	sedaPayloadHash := hasher.Sum(nil)

	hasher.Reset()
	allBytes := make([]byte, 0, len(versionHash)+len(drIDBytes)+1+1+len(resultHash)+len(blockHeightBytes)+len(blockTimestampBytes)+len(gasUsedBytes)+len(paybackAddrHash)+len(sedaPayloadHash))
	allBytes = append(allBytes, versionHash...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, consensusByte)
	allBytes = append(allBytes, exitCodeByte)
	allBytes = append(allBytes, resultHash...)
	allBytes = append(allBytes, blockHeightBytes...)
	allBytes = append(allBytes, blockTimestampBytes...)
	allBytes = append(allBytes, gasUsedBytes...)
	allBytes = append(allBytes, paybackAddrHash...)
	allBytes = append(allBytes, sedaPayloadHash...)
	hasher.Write(allBytes)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
