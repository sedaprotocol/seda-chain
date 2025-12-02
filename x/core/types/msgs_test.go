package types

import (
	"encoding/hex"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	"cosmossdk.io/math"
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
				ExecProgramID:   validProgramID,
				TallyProgramID:  validProgramID,
				ExecInputs:      make([]byte, 100),
				TallyInputs:     make([]byte, 100),
				ConsensusFilter: make([]byte, 100),
				Memo:            make([]byte, 100),
				PaybackAddress:  make([]byte, 100),
				SEDAPayload:     make([]byte, 100),
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:  validProgramID,
				TallyProgramID: validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdefa",
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdzz",
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID + "ab",
			},
			config:  *validConfig,
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
				ExecProgramID:     validProgramID,
				TallyProgramID:    validProgramID,
			},
			config:  *validConfig,
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

func TestAddToAllowlist_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgAddToAllowlist
		wantErr error
	}{
		{
			name: "valid message",
			msg: MsgAddToAllowlist{
				Sender:    "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f",
				PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
			},
			wantErr: nil,
		},
		{
			name: "invalid sender",
			msg: MsgAddToAllowlist{
				Sender:    "seda1invalid",
				PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid public key",
			msg: MsgAddToAllowlist{
				Sender:    "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f",
				PublicKey: "invalid",
			},
			wantErr: ErrInvalidStakerPublicKey,
		},
		{
			name: "empty public key",
			msg: MsgAddToAllowlist{
				Sender:    "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f",
				PublicKey: "",
			},
			wantErr: ErrInvalidStakerPublicKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestStakeProof tests against a given stake proof.
func TestStakeProof(t *testing.T) {
	chainID := "seda-1-devnet"
	seqNum := uint64(0)
	msg := MsgStake{
		PublicKey: "02bd3aeeea42da249900c97cb63f1647e8c797a432e3b508dcd3e8f70a89b1a622",
		Memo:      "c2VkYXByb3RvY29s",
		Proof:     "030b3a90682d42547987d283027f71e5b434087ae6fd2c46d2cccc870d8b90ca71e0b8331e8467028665341f63a21765c0f4687798aa13b8655fe170f7d7132959e37ae35a2a7e0664a87eb6345c06d507",
	}

	publicKey, err := hex.DecodeString(msg.PublicKey)
	require.NoError(t, err)
	proof, err := hex.DecodeString(msg.Proof)
	require.NoError(t, err)

	_, err = vrf.NewK256VRF().Verify(publicKey, proof, msg.MsgHash("", chainID, seqNum))
	require.NoError(t, err)
}

// TestProveAndVerifyStakeProof produces a proof and verifies it.
func TestProveAndVerifyStakeProof(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().Bytes()
	chainID := "seda-1-devnet"
	msg := MsgStake{
		PublicKey: hex.EncodeToString(pubKey),
		Memo:      "VGhlIFNpbmdsZSBVTklYIFNwZWNpZmljYXRpb24gc3VwcG9ydHMgZm9ybWFsIHN0YW5kYXJkcyBkZXZlbG9wZWQgZm9yIGFwcGxpY2F0aW9ucyBwb3J0YWJpbGl0eS4g",
	}
	hash := msg.MsgHash("", chainID, 99)

	proof, err := vrf.NewK256VRF().Prove(privKey, hash)
	require.NoError(t, err)

	_, err = vrf.NewK256VRF().Verify(pubKey, proof, hash)
	require.NoError(t, err)
}

// Note seda-overlay-ts has the same test case.
func TestCommitMsgHash(t *testing.T) {
	chainID := "seda_test"
	msg := MsgCommit{
		DrID:   "3aa91e148d735de527a185f5ff36238dc4edae93605a1e0bb09962a2f64a818f",
		Commit: "5cd42fbed0f93a8ad51098c3c3354203acbfd59ba67f0b1304a8db4938f4cba9",
	}
	require.Equal(t, "cea5308a78283be4d02bae4db034680995815d7371caa2f034397cfa15baf554", hex.EncodeToString(msg.MsgHash("", chainID, 1)))
}

// Note seda-overlay-ts has the same test case.
func TestRevealMsgHash(t *testing.T) {
	chainID := "seda_test"
	reveal, err := hex.DecodeString("ccb1f717aa77602faf03a594761a36956b1c4cf44c6b336d1db57da799b331b8")
	require.NoError(t, err)
	privKey, err := hex.DecodeString("2bd806c97f0e00af1a1fc3328fa763a9269723c8db8fac4f93af71db186d6e90")
	require.NoError(t, err)

	revealMsg := MsgReveal{
		RevealBody: &RevealBody{
			DrID:          "3aa91e148d735de527a185f5ff36238dc4edae93605a1e0bb09962a2f64a818f",
			DrBlockHeight: 1,
			ExitCode:      0,
			GasUsed:       1,
			Reveal:        reveal,
			ProxyPubKeys: []string{
				"030123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			},
		},
		PublicKey: "039997a497d964fc1a62885b05a51166a65a90df00492c8d7cf61d6accf54803be",
	}
	revealMsgHash := revealMsg.MsgHash("", chainID)
	require.Equal(t, "e4c4ce71f72a0b69c0b0cf06eed6de1cdae2998a9e5d387824ae7f7ebe93f6db", hex.EncodeToString(revealMsgHash))

	revealProof, err := vrf.NewK256VRF().Prove(privKey, revealMsgHash)
	require.NoError(t, err)
	require.Equal(t, "027dba0119599c4f13818db05ba3cb62c0d1eafcb00df4c76437a5bf31a79670de7a3604f228d24c36cc41d991711eb739dbc8054e597b37402a0c45ad123cd0bb214716a55d013eba5776052a72fe1335", hex.EncodeToString(revealProof))

	revealMsg.Proof = hex.EncodeToString(revealProof)
	revealHash := revealMsg.RevealHash()
	require.NoError(t, err)
	require.Equal(t, "85cc1478cc060f15edcbd7a89fd61b7d6056d243eddbcef261fd6f05a4054cc9", hex.EncodeToString(revealHash))

	commitMsg := MsgCommit{
		DrID:   revealMsg.RevealBody.DrID,
		Commit: hex.EncodeToString(revealHash),
	}
	commitMsgHash := commitMsg.MsgHash("", chainID, int64(revealMsg.RevealBody.DrBlockHeight))
	require.Equal(t, "45d1e862dd1a929bc91befe81e1db3c70ad19bca9c32fcfffdd5e2812c2ddb55", hex.EncodeToString(commitMsgHash))

	commitProof, err := vrf.NewK256VRF().Prove(privKey, commitMsgHash)
	require.NoError(t, err)
	require.Equal(t, "037af31e9a2b73049123eb3cf2d6cfb0a0aa221673fbcbcdbd3f6a10333f0e47cf09dc53fb1c35d2b573feb5840dec8d830d378b874d8fea644da20c71ebb0598236936aca56704730363cc7708245f32b", hex.EncodeToString(commitProof))
}
