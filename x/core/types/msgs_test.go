package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestMsgPostDataRequest_Validate(t *testing.T) {
	validProgramID := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	validConfig := DefaultParams().DataRequestConfig
	validSender := "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f"
	validVersion := "1.0.0"

	tests := []struct {
		name    string
		msg     MsgPostDataRequest
		config  DataRequestConfig
		wantErr error
	}{
		{
			name: "valid message",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 65535, // max uint16 value
				GasPrice: func() math.Int {
					max, ok := math.NewIntFromString("340282366920938463463374607431768211455") // max uint128 value
					require.True(t, ok)
					return max
				}(),
				Version:         validVersion,
				ExecGasLimit:    MinExecGasLimit,
				TallyGasLimit:   MinTallyGasLimit,
				ExecProgramId:   validProgramID,
				TallyProgramId:  validProgramID,
				ExecInputs:      make([]byte, 100),
				TallyInputs:     make([]byte, 100),
				ConsensusFilter: make([]byte, 100),
				Memo:            make([]byte, 100),
				PaybackAddress:  make([]byte, 100),
				SEDAPayload:     make([]byte, 100),
			},
			config:  validConfig,
			wantErr: nil,
		},
		{
			name: "zero replication factor",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 0,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrZeroReplicationFactor,
		},
		{
			name: "replication factor too high (exceeds uint16)",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 65536, // exceeds uint16
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrReplicationFactorNotUint16,
		},
		{
			name: "gas price too low",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice.Sub(math.NewInt(1)),
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrGasPriceTooLow,
		},
		{
			name: "gas price too high (exceeds uint128)",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice: func() math.Int {
					max, ok := math.NewIntFromString("340282366920938463463374607431768211456")
					require.True(t, ok)
					return max
				}(),
				Version:        validVersion,
				ExecGasLimit:   MinExecGasLimit,
				TallyGasLimit:  MinTallyGasLimit,
				ExecProgramId:  validProgramID,
				TallyProgramId: validProgramID,
			},
			config:  validConfig,
			wantErr: ErrGasPriceTooHigh,
		},
		{
			name: "exec gas limit too low",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit - 1,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrExecGasLimitTooLow,
		},
		{
			name: "tally gas limit too low",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit - 1,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrTallyGasLimitTooLow,
		},
		{
			name: "exec program ID is not hex (odd length)",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdefa",
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrInvalidExecProgramID,
		},
		{
			name: "tally program ID is not hex",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdzz",
			},
			config:  validConfig,
			wantErr: ErrInvalidTallyProgramID,
		},
		{
			name: "tally program ID too long",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           validVersion,
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID + "ab",
			},
			config:  validConfig,
			wantErr: ErrInvalidLengthTallyProgramID,
		},
		{
			name: "invalid version",
			msg: MsgPostDataRequest{
				Sender:            validSender,
				ReplicationFactor: 5,
				GasPrice:          MinGasPrice,
				Version:           "1.0.0-alpha.1",
				ExecGasLimit:      MinExecGasLimit,
				TallyGasLimit:     MinTallyGasLimit,
				ExecProgramId:     validProgramID,
				TallyProgramId:    validProgramID,
			},
			config:  validConfig,
			wantErr: ErrInvalidVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate(tt.config)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
