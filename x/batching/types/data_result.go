package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"

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

	exitCodeByte := byte(dr.ExitCode)

	hasher.Reset()
	hasher.Write(dr.Result)
	resultHash := hasher.Sum(nil)

	gasUsedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasUsedBytes, dr.GasUsed)

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
	var allBytes []byte
	allBytes = append(allBytes, versionHash...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, consensusByte)
	allBytes = append(allBytes, exitCodeByte)
	allBytes = append(allBytes, resultHash...)
	allBytes = append(allBytes, blockHeightBytes...)
	allBytes = append(allBytes, gasUsedBytes...)
	allBytes = append(allBytes, paybackAddrHash...)
	allBytes = append(allBytes, sedaPayloadHash...)
	hasher.Write(allBytes)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
