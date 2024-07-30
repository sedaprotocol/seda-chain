package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	vrf "github.com/sedaprotocol/vrf-go"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const VRFKeyFileName = "vrf_key.json"

type VRFKey struct {
	Address types.Address    `json:"address"`
	PubKey  sdkcrypto.PubKey `json:"pub_key"`
	PrivKey crypto.PrivKey   `json:"priv_key"`

	filePath string
	vrf      *vrf.VRFStruct
}

// Save persists the VRFKey to its filePath.
func (v VRFKey) Save() error {
	outFile := v.filePath
	if outFile == "" {
		return fmt.Errorf("key's file path is empty")
	}

	cmtPubKey, err := cryptocodec.ToCmtPubKeyInterface(v.PubKey)
	if err != nil {
		return fmt.Errorf("failed to convert key type from SDK to Comet: %v", err)
	}

	vrfKeyFile := struct {
		PrivKey crypto.PrivKey `json:"priv_key"`
		PubKey  crypto.PubKey  `json:"pub_key"`
	}{
		PrivKey: v.PrivKey,
		PubKey:  cmtPubKey,
	}

	jsonBytes, err := cmtjson.MarshalIndent(vrfKeyFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal key: %v", err)
	}

	err = os.WriteFile(outFile, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	return nil
}

// VRFProve uses the VRF key to compute the VRF hash output (beta)
// and the proof that it was computed correctly (pi).
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

// VRFVerify verifies that beta is the correct VRF hash of the alpha
// under private key associated with the given public key. It also
// outputs the hash output beta.
func (v *VRFKey) VRFVerify(publicKey, alpha, pi []byte) (beta []byte, err error) {
	beta, err = v.vrf.Verify(publicKey, alpha, pi)
	if err != nil {
		return nil, err
	}
	return beta, nil
}

// SignTransaction signs a given transaction with the VRF key and
// returns the resulting signature. The given account must belong
// to the VRF key.
func (v *VRFKey) SignTransaction(
	ctx sdk.Context, txBuilder client.TxBuilder, txConfig client.TxConfig,
	signMode txsigning.SignMode, account sdk.AccountI,
) (txsigning.SignatureV2, error) {
	var sigV2 txsigning.SignatureV2

	if !bytes.Equal(account.GetPubKey().Bytes(), v.PubKey.Bytes()) {
		return sigV2, fmt.Errorf("the account does not belong to the vrf key")
	}

	signerData := authsigning.SignerData{
		ChainID:       ctx.ChainID(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
		PubKey:        v.PubKey,
		Address:       account.GetAddress().String(),
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	nilSig := txsigning.SignatureV2{
		PubKey: v.PubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  signMode,
			Signature: nil,
		},
		Sequence: account.GetSequence(),
	}

	if err := txBuilder.SetSignatures(nilSig); err != nil {
		return sigV2, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		ctx,
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

	sigV2 = txsigning.SignatureV2{
		PubKey: v.PubKey,
		Data: &txsigning.SingleSignatureData{
			SignMode:  signMode,
			Signature: sigBytes,
		},
		Sequence: account.GetSequence(),
	}
	return sigV2, nil
}

func (v *VRFKey) IsNil() bool {
	return v == nil
}

// NewVRFKey generates a new VRFKey from the given key and key file path.
func NewVRFKey(privKey crypto.PrivKey, keyFilePath string) (*VRFKey, error) {
	vrfStruct := vrf.NewK256VRF()
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

func LoadVRFKey(keyFilePath string) (*VRFKey, error) {
	keyJSONBytes, err := os.ReadFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading VRF key from %v: %v", keyFilePath, err)
	}

	vrfKeyFile := struct {
		PrivKey crypto.PrivKey `json:"priv_key"`
	}{}
	err = cmtjson.Unmarshal(keyJSONBytes, &vrfKeyFile)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling VRF key from %v: %v", keyFilePath, err)
	}

	vrfKey, err := NewVRFKey(vrfKeyFile.PrivKey, keyFilePath)
	if err != nil {
		return nil, err
	}

	return vrfKey, nil
}

// LoadOrGenVRFKey initializes a VRF key and returns its public key.
// If keyFilePath is specified, it loads the VRF key file at the specified
// path. Otherwise, it generates a new VRF key, whose entropy is randomly
// generated or obtained from the mnemonic, if provided.
func LoadOrGenVRFKey(config *cfg.Config, keyFilePath, mnemonic string) (vrfPubKey sdkcrypto.PubKey, err error) {
	var vrfKey *VRFKey
	if keyFilePath != "" {
		vrfKey, err = LoadVRFKey(keyFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		var privKey secp256k1.PrivKey
		if mnemonic != "" {
			privKey = secp256k1.GenPrivKeySecp256k1([]byte(mnemonic))
		} else {
			privKey = secp256k1.GenPrivKey()
		}

		// VRF key file is placed in the same directory as the validator key file.
		keyFilePath := filepath.Join(filepath.Dir(config.PrivValidatorKeyFile()), VRFKeyFileName)
		if cmtos.FileExists(keyFilePath) {
			return nil, fmt.Errorf("vrf key file already exists at %s", keyFilePath)
		}
		vrfKey, err = NewVRFKey(privKey, keyFilePath)
		if err != nil {
			return nil, err
		}
		err = vrfKey.Save()
		if err != nil {
			return nil, err
		}
	}
	return vrfKey.PubKey, nil
}
