package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvidenceValidateBasic(t *testing.T) {
	tests := []struct {
		name     string
		evidence *BatchDoubleSign
		wantErr  error
	}{
		{
			name: "Happy path",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: nil,
		},
		{
			name: "Invalid batch number",
			evidence: &BatchDoubleSign{
				BatchNumber:         0,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: fmt.Errorf("batch number must be greater than 1"),
		},
		{
			name: "Invalid block height",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         -100,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: fmt.Errorf("invalid block height: -100"),
		},
		{
			name: "Invalid validator operator address",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: fmt.Errorf("invalid operator address: "),
		},
		{
			name: "Invalid validator root",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: fmt.Errorf("invalid validator root: "),
		},
		{
			name: "Invalid data result root",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: fmt.Errorf("invalid data result root: "),
		},
		{
			name: "Invalid proving metadata hash",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "",
			},
			wantErr: fmt.Errorf("invalid proving metadata hash: "),
		},
		{
			name: "Invalid proving scheme index",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
				// Seems unlikely that this will ever be used
				ProvingSchemeIndex: 999999,
			},
			wantErr: fmt.Errorf("invalid proving scheme index: 999999"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(ti *testing.T) {
			err := tt.evidence.ValidateBasic()
			if tt.wantErr != nil {
				require.Error(ti, err)
				require.Equal(ti, tt.wantErr, err)
			} else {
				require.NoError(ti, err)
			}
		})
	}
}

func TestEvidenceBatchID(t *testing.T) {
	tests := []struct {
		name        string
		evidence    *BatchDoubleSign
		wantBatchID string
		wantErr     error
	}{
		{
			name: "Happy path",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantBatchID: "7e263196439561a12d3488a44302605ca27df1ecc4d5e07e42f3630a8435ae88",
			wantErr:     nil,
		},
		{
			name: "Invalid validator root",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "g027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: hex.InvalidByteError('g'),
		},
		{
			name: "Invalid data result root",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "z306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: hex.InvalidByteError('z'),
		},
		{
			name: "Invalid proving metadata hash",
			evidence: &BatchDoubleSign{
				BatchNumber:         2,
				BlockHeight:         48,
				OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
				ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
				DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
				ProvingMetadataHash: "p000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: hex.InvalidByteError('p'),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(ti *testing.T) {
			batchID, err := tt.evidence.GetBatchID()
			if tt.wantErr != nil {
				require.Error(ti, err)
				require.Equal(ti, tt.wantErr, err)
			} else {
				require.NoError(ti, err)
				require.Equal(ti, tt.wantBatchID, hex.EncodeToString(batchID))
			}
		})
	}
}

// Since the evidence module relies of the hash of the evidence to be unique we add this test
// to be notified if the hash changes.
func TestEvidenceHash(t *testing.T) {
	evidence := &BatchDoubleSign{
		BatchNumber:         2,
		BlockHeight:         48,
		OperatorAddress:     "sedavaloper1easwjglg0l6s5qrnlhqd25l2x0h5gy7law835s",
		ValidatorRoot:       "6027c97e8b0588f86a9e140d73a31af5ee0d37b93ff0f2f54f5305d0f2ea3fd9",
		DataResultRoot:      "2306d94cc69db8435c56294ff7f27cf3a7d042f8965e2d76f38c63a616a937b0",
		ProvingMetadataHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}
	hash := evidence.Hash()

	require.Equal(t, "986aa7a4d4c0213cc7a2ca53807af3a79abdaf47ab55c36d96067776721cd5e2", hex.EncodeToString(hash), "If this test fails it means that old evidence could be resubmitted!")
}
