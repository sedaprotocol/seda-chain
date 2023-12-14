package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	vrf "github.com/sedaprotocol/vrf-go"
)

var _ VRFSigner = &VRFKey{}

type VRFSigner interface {
	VRFProve(alpha []byte) (pi, beta []byte, err error)
	SignTransaction(ctx sdk.Context,
		signMode signing.SignMode,
		txBuilder client.TxBuilder, txConfig client.TxConfig, account sdk.AccountI) (signing.SignatureV2, error)
}

type VRFKey struct {
	Address types.Address    `json:"address"`
	PubKey  sdkcrypto.PubKey `json:"pub_key"`
	PrivKey crypto.PrivKey   `json:"priv_key"` // TO-DO can we not export it?

	filePath string
	vrf      *vrf.VRFStruct
}

// Save persists the VRFKey to its filePath.
func (key VRFKey) Save() error {
	outFile := key.filePath
	if outFile == "" {
		return fmt.Errorf("key's file path is empty")
	}

	jsonBytes, err := cmtjson.MarshalIndent(key, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key: %v", err)
	}

	err = os.WriteFile(outFile, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	return nil
}

func (v *VRFKey) VRFProve(alpha []byte) (pi, beta []byte, err error) {
	pi, err = v.vrf.Prove(v.PrivKey.Bytes(), alpha)
	if err != nil {
		return nil, nil, err
	}
	beta, err = v.vrf.ProofToHash(pi)
	if err != nil {
		return nil, nil, err
	}
	return pi, beta, nil
}

func (v *VRFKey) SignTransaction(
	ctx sdk.Context,
	signMode signing.SignMode,
	txBuilder client.TxBuilder, txConfig client.TxConfig, account sdk.AccountI,
) (signing.SignatureV2, error) {
	var sigV2 signing.SignatureV2

	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
		PubKey:        v.PubKey,
		Address:       account.GetAddress().String(),
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signMode,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return sigV2, err
	}

	sigBytes, err := v.PrivKey.Sign(bytesToSign)
	if err != nil {
		return sigV2, err
	}

	sigV2 = signing.SignatureV2{
		PubKey: v.PubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  signMode,
			Signature: sigBytes,
		},
		Sequence: account.GetSequence(),
	}

	return sigV2, nil
}

// NewVRFKey generates a new VRFKey from the given key and key file path.
func NewVRFKey(privKey crypto.PrivKey, keyFilePath string) (*VRFKey, error) {
	vrfStruct := vrf.NewK256VRF(0xFE)
	pubKey, err := cryptocodec.FromCmtPubKeyInterface(privKey.PubKey())
	if err != nil {
		return nil, err
	}
	return &VRFKey{
		Address:  privKey.PubKey().Address(),
		PubKey:   pubKey,
		PrivKey:  privKey,
		filePath: keyFilePath,
		vrf:      &vrfStruct,
	}, nil
}

// GenVRFKey generates a new VRFKey with a randomly generated private key.
func GenVRFKey(keyFilePath string) (*VRFKey, error) {
	return NewVRFKey(secp256k1.GenPrivKey(), keyFilePath)
}

func LoadVRFKey(keyFilePath string) (*VRFKey, error) {
	keyJSONBytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading VRF key from %v: %v", keyFilePath, err)
	}
	vrfKey := new(VRFKey)
	err = cmtjson.Unmarshal(keyJSONBytes, vrfKey)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling VRF key from %v: %v", keyFilePath, err)
	}
	return vrfKey, nil
}

// LoadOrGenVRFKey loads a VRFKey from the given file path
// or else generates a new one and saves it to the file path.
func LoadOrGenVRFKey(keyFilePath string) (*VRFKey, error) {
	var vrfKey *VRFKey
	var err error
	if cmtos.FileExists(keyFilePath) {
		vrfKey, err = LoadVRFKey(keyFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		vrfKey, err = GenVRFKey(keyFilePath)
		if err != nil {
			return nil, err
		}
		err = vrfKey.Save()
		if err != nil {
			return nil, err
		}
	}
	return vrfKey, nil
}
