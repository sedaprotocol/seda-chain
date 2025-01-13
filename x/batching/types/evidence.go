package types

import (
	"encoding/hex"
	"fmt"

	"github.com/cometbft/cometbft/crypto/tmhash"

	"cosmossdk.io/x/evidence/exported"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

// Evidence type constants
const RouteBatchDoubleSign = "batchdoublesign"

var _ exported.Evidence = &BatchDoubleSign{}

// Route returns the Evidence Handler route for a BatchDoubleSign type.
func (e *BatchDoubleSign) Route() string { return RouteBatchDoubleSign }

// Hash returns the hash of a BatchDoubleSign object.
func (e *BatchDoubleSign) Hash() []byte {
	bz, err := e.Marshal()
	if err != nil {
		panic(err)
	}
	return tmhash.Sum(bz)
}

// ValidateBasic performs basic stateless validation checks on a BatchDoubleSign object.
func (e *BatchDoubleSign) ValidateBasic() error {
	if e.BatchNumber <= 1 {
		return fmt.Errorf("batch number must be greater than 1")
	}

	if e.BlockHeight < 1 {
		return fmt.Errorf("invalid block height: %d", e.BlockHeight)
	}

	if e.OperatorAddress == "" {
		return fmt.Errorf("invalid operator address: %s", e.OperatorAddress)
	}

	if e.ValidatorRoot == "" {
		return fmt.Errorf("invalid validator root: %s", e.ValidatorRoot)
	}

	if e.DataResultRoot == "" {
		return fmt.Errorf("invalid data result root: %s", e.DataResultRoot)
	}

	if e.ProvingMetadataHash == "" {
		return fmt.Errorf("invalid proving metadata hash: %s", e.ProvingMetadataHash)
	}

	// Currently we only support secp256k1 signatures for batch signing
	if e.ProvingSchemeIndex != uint32(utils.SEDAKeyIndexSecp256k1) {
		return fmt.Errorf("invalid proving scheme index: %d", e.ProvingSchemeIndex)
	}

	return nil
}

func (e *BatchDoubleSign) GetBatchID() ([]byte, error) {
	validatorRoot, err := hex.DecodeString(e.ValidatorRoot)
	if err != nil {
		return nil, err
	}

	dataResultRoot, err := hex.DecodeString(e.DataResultRoot)
	if err != nil {
		return nil, err
	}

	provingMetadataHash, err := hex.DecodeString(e.ProvingMetadataHash)
	if err != nil {
		return nil, err
	}

	return ComputeBatchID(e.BatchNumber, e.BlockHeight, validatorRoot, dataResultRoot, provingMetadataHash), nil
}

// GetHeight returns the height at time of the BatchDoubleSign infraction.
func (e *BatchDoubleSign) GetHeight() int64 {
	return e.BlockHeight
}
