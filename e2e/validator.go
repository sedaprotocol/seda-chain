package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	tmcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"

	"cosmossdk.io/math"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/utils"
	sedastakingtypes "github.com/sedaprotocol/seda-chain/x/staking/types"
)

type validator struct {
	chain            *chain
	index            int
	moniker          string
	mnemonic         string
	keyInfo          keyring.Record
	privateKey       cryptotypes.PrivKey
	consensusKey     privval.FilePVKey
	consensusPrivKey cryptotypes.PrivKey //nolint:unused // unused
	nodeKey          p2p.NodeKey
}

type account struct {
	moniker    string //nolint:unused // unused
	mnemonic   string
	keyInfo    keyring.Record
	privateKey cryptotypes.PrivKey
}

func (v *validator) instanceName() string {
	return fmt.Sprintf("%s%d", v.moniker, v.index)
}

func (v *validator) configDir() string {
	return fmt.Sprintf("%s/%s", v.chain.configDir(), v.instanceName())
}

func (v *validator) createConfig() error {
	p := path.Join(v.configDir(), "config")
	return os.MkdirAll(p, 0o700)
}

func (v *validator) init() error {
	if err := v.createConfig(); err != nil {
		return err
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	appState, err := json.MarshalIndent(app.ModuleBasics.DefaultGenesis(cdc), "", " ")
	if err != nil {
		return fmt.Errorf("failed to JSON encode app genesis state: %w", err)
	}

	appGenesis := genutiltypes.NewAppGenesisWithVersion(v.chain.id, appState)

	err = appGenesis.SaveAs(config.GenesisFile())
	if err != nil {
		return err
	}

	tmcfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	return nil
}

func (v *validator) createNodeKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	v.nodeKey = *nodeKey
	return nil
}

func (v *validator) createConsensusKey() error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(v.configDir())
	config.Moniker = v.moniker

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := tmos.EnsureDir(filepath.Dir(pvKeyFile), 0o700); err != nil {
		return err
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := tmos.EnsureDir(filepath.Dir(pvStateFile), 0o700); err != nil {
		return err
	}

	filePV := privval.NewFilePV(ed25519.GenPrivKey(), config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	filePV.Save()

	v.consensusKey = filePV.Key
	return nil
}

func (v *validator) createKeyFromMnemonic(name, mnemonic string) error {
	dir := v.configDir()
	kb, err := keyring.New(keyringAppName, keyring.BackendTest, dir, nil, cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return err
	}

	privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
	if err != nil {
		return err
	}

	privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
	if err != nil {
		return err
	}

	v.keyInfo = *info
	v.mnemonic = mnemonic
	v.privateKey = privKey

	return nil
}

func (c *chain) addAccountFromMnemonic(counts int) error {
	val0ConfigDir := c.validators[0].configDir()
	kb, err := keyring.New(keyringAppName, keyring.BackendTest, val0ConfigDir, nil, cdc)
	if err != nil {
		return err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return err
	}

	for i := 0; i < counts; i++ {
		name := fmt.Sprintf("acct-%d", i)
		mnemonic, err := createMnemonic()
		if err != nil {
			return err
		}
		info, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
		if err != nil {
			return err
		}

		privKeyArmor, err := kb.ExportPrivKeyArmor(name, keyringPassphrase)
		if err != nil {
			return err
		}

		privKey, _, err := sdkcrypto.UnarmorDecryptPrivKey(privKeyArmor, keyringPassphrase)
		if err != nil {
			return err
		}
		acct := account{}
		acct.keyInfo = *info
		acct.mnemonic = mnemonic
		acct.privateKey = privKey
		c.genesisAccounts = append(c.genesisAccounts, &acct)
	}

	return nil
}

func (v *validator) createKey(name string) error {
	mnemonic, err := createMnemonic()
	if err != nil {
		return err
	}

	return v.createKeyFromMnemonic(name, mnemonic)
}

func (v *validator) buildCreateValidatorMsg(amount sdk.Coin, valHomeDir string) (sdk.Msg, error) {
	description := stakingtypes.NewDescription(v.moniker, "", "", "", "")
	commissionRates := stakingtypes.CommissionRates{
		Rate:          math.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       math.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: math.LegacyMustNewDecFromStr("0.01"),
	}

	valPubKey, err := cryptocodec.FromCmtPubKeyInterface(v.consensusKey.PubKey)
	if err != nil {
		return nil, err
	}

	valAddr, err := v.keyInfo.GetAddress()
	if err != nil {
		return nil, err
	}

	sedaPubKeys, err := utils.GenerateSEDAKeys(sdk.ValAddress(valAddr), filepath.Join(valHomeDir, "config/seda_keys.json"), "", true)
	if err != nil {
		return nil, err
	}

	return sedastakingtypes.NewMsgCreateSEDAValidator(
		sdk.ValAddress(valAddr).String(),
		valPubKey,
		sedaPubKeys,
		amount,
		description,
		commissionRates,
		math.NewInt(1),
	)
}

func (v *validator) signMsg(msgs ...sdk.Msg) (*sdktx.Tx, error) {
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetMemo(fmt.Sprintf("%s@%s:26656", v.nodeKey.ID(), v.instanceName()))
	txBuilder.SetFeeAmount(sdk.NewCoins())
	txBuilder.SetGasLimit(200000)

	pk, err := v.keyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}

	signerData := authsigning.SignerData{
		ChainID:       v.chain.id,
		AccountNumber: 0,
		Sequence:      0,
		PubKey:        pk,
		Address:       sdk.AccAddress(pk.Address()).String(),
	}

	// For SIGN_MODE_DIRECT, calling SetSignatures calls setSignerInfos on
	// TxBuilder under the hood, and SignerInfos is needed to generate the sign
	// bytes. This is the reason for setting SetSignatures here, with a nil
	// signature.
	//
	// Note: This line is not needed for SIGN_MODE_LEGACY_AMINO, but putting it
	// also doesn't affect its generated sign bytes, so for code's simplicity
	// sake, we put it here.
	sig := txsigning.SignatureV2{
		PubKey: pk,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: 0,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	bytesToSign, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		encodingConfig.TxConfig.SignModeHandler(),
		txsigning.SignMode_SIGN_MODE_DIRECT,
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, err
	}

	sigBytes, err := v.privateKey.Sign(bytesToSign)
	if err != nil {
		return nil, err
	}

	pk, err = v.keyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}

	sig = txsigning.SignatureV2{
		PubKey: pk,
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: sigBytes,
		},
		Sequence: 0,
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, err
	}

	signedTx := txBuilder.GetTx()
	bz, err := encodingConfig.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, err
	}

	return decodeTx(bz)
}
