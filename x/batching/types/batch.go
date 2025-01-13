package types

import (
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

// Computes the batch ID, which is defined as
// keccak256(batch_number, block_height, validator_root, results_root, proving_metadata_hash)
func ComputeBatchID(batchNumber uint64, blockHeight int64, validatorRoot []byte, dataResultRoot []byte, provingMetadataHash []byte) []byte {
	var hashContent []byte
	hashContent = binary.BigEndian.AppendUint64(hashContent, batchNumber)
	//nolint:gosec // G115: We shouldn't get negative block heights anyway.
	hashContent = binary.BigEndian.AppendUint64(hashContent, uint64(blockHeight))
	hashContent = append(hashContent, validatorRoot...)
	hashContent = append(hashContent, dataResultRoot...)
	hashContent = append(hashContent, provingMetadataHash...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(hashContent)
	batchID := hasher.Sum(nil)

	return batchID
}
