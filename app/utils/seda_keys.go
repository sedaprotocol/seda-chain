package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	cmtos "github.com/cometbft/cometbft/libs/os"

	sdk "github.com/cosmos/cosmos-sdk/types"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// SEDAKeyIndex enumerates the SEDA key indices.
type SEDAKeyIndex uint32

const (
	SEDAKeyIndexSecp256k1 SEDAKeyIndex = iota
)

// SEDA domain separators
const (
	SEDASeparatorDataRequest byte = iota
	SEDASeparatorSecp256k1
)

type privKeyGenerator func() *ecdsa.PrivateKey

// sedaKeyGenerators maps the SEDA key index to the corresponding
// private key generator.
var sedaKeyGenerators = map[SEDAKeyIndex]privKeyGenerator{
	SEDAKeyIndexSecp256k1: func() *ecdsa.PrivateKey {
		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		return privKey
	},
}

type pubKeyValidator func([]byte) bool

// sedaPubKeyValidators maps the SEDA key index to the corresponding
// public key validator.
var sedaPubKeyValidators = map[SEDAKeyIndex]pubKeyValidator{
	SEDAKeyIndexSecp256k1: func(pub []byte) bool {
		_, err := ethcrypto.UnmarshalPubkey(pub)
		return err == nil
	},
}

// SEDAKeyFileName defines the SEDA key file name.
const SEDAKeyFileName = "seda_keys.json"

type sedaKeyFile struct {
	ValidatorAddr sdk.ValAddress   `json:"validator_addr"`
	Keys          []indexedPrivKey `json:"keys"`
}

// indexedPrivKey is used for persisting the SEDA keys in a file.
type indexedPrivKey struct {
	Index   SEDAKeyIndex      `json:"index"`
	PrivKey *ecdsa.PrivateKey `json:"priv_key"`
	PubKey  []byte            `json:"pub_key"`
}

func (k *indexedPrivKey) MarshalJSON() ([]byte, error) {
	type Alias indexedPrivKey
	return json.Marshal(&struct {
		*Alias
		PrivKey string `json:"priv_key"`
	}{
		Alias:   (*Alias)(k),
		PrivKey: fmt.Sprintf("%x", ethcrypto.FromECDSA(k.PrivKey)),
	})
}

func (k *indexedPrivKey) UnmarshalJSON(data []byte) error {
	type Alias indexedPrivKey
	aux := &struct {
		*Alias
		PrivKey string `json:"priv_key"`
	}{
		Alias: (*Alias)(k),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	privBytes, err := hex.DecodeString(aux.PrivKey)
	if err != nil {
		return fmt.Errorf("failed to decode private key hex: %v", err)
	}
	k.PrivKey, err = ethcrypto.ToECDSA(privBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}
	return nil
}

// saveSEDAKeys saves a given list of indexedPrivKey in the directory
// at dirPath.
func saveSEDAKeys(keys []indexedPrivKey, valAddr sdk.ValAddress, dirPath string) error {
	savePath := filepath.Join(dirPath, SEDAKeyFileName)
	if cmtos.FileExists(savePath) {
		return fmt.Errorf("SEDA key file already exists at %s", savePath)
	}
	err := cmtos.EnsureDir(filepath.Dir(savePath), 0o700)
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(sedaKeyFile{
		ValidatorAddr: valAddr,
		Keys:          keys,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SEDA keys: %v", err)
	}

	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write SEDA key file: %v", err)
	}
	return nil
}

// loadSEDAKeys loads the SEDA key file from the given path.
func loadSEDAKeys(loadPath string) (sedaKeyFile, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return sedaKeyFile{}, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keyFile sedaKeyFile
	err = json.Unmarshal(keysJSONBytes, &keyFile)
	if err != nil {
		return sedaKeyFile{}, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}
	return keyFile, nil
}

// LoadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs.
func LoadSEDAPubKeys(loadPath string) ([]pubkeytypes.IndexedPubKey, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keyFile sedaKeyFile
	err = json.Unmarshal(keysJSONBytes, &keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	result := make([]pubkeytypes.IndexedPubKey, len(keyFile.Keys))
	for i, key := range keyFile.Keys {
		pubKey := key.PrivKey.PublicKey
		pubKeyBytes := ethcrypto.FromECDSAPub(&pubKey)
		result[i] = pubkeytypes.IndexedPubKey{
			Index:  uint32(key.Index),
			PubKey: pubKeyBytes,
		}
	}
	return result, nil
}

// GenerateSEDAKeys generates a new set of SEDA keys and saves them to
// the SEDA key file, along with the provided validator address. It
// returns the resulting index-public key pairs. The key file is stored
// in the directory given by dirPath.
func GenerateSEDAKeys(valAddr sdk.ValAddress, dirPath string) ([]pubkeytypes.IndexedPubKey, error) {
	privKeys := make([]indexedPrivKey, 0, len(sedaKeyGenerators))
	pubKeys := make([]pubkeytypes.IndexedPubKey, 0, len(sedaKeyGenerators))
	for keyIndex, generator := range sedaKeyGenerators {
		privKey := generator()
		pubKey := ethcrypto.FromECDSAPub(&privKey.PublicKey)

		privKeys = append(privKeys, indexedPrivKey{
			Index:   keyIndex,
			PrivKey: privKey,
			PubKey:  pubKey,
		})
		pubKeys = append(pubKeys, pubkeytypes.IndexedPubKey{
			Index:  uint32(keyIndex),
			PubKey: pubKey,
		})
	}

	// The key file is placed in the same directory as the validator key file.
	err := saveSEDAKeys(privKeys, valAddr, dirPath)
	if err != nil {
		return nil, err
	}
	return pubKeys, nil
}

// ValidateSEDAPubKeys ensures that the provided indexed public keys
// conform to SEDA keys specifications. It first sorts the provided
// slice for deterministic results.
func ValidateSEDAPubKeys(indPubKeys []pubkeytypes.IndexedPubKey) error {
	if len(sedaPubKeyValidators) != len(indPubKeys) {
		return fmt.Errorf("invalid number of SEDA keys")
	}
	sort.Slice(indPubKeys, func(i, j int) bool {
		return indPubKeys[i].Index < indPubKeys[j].Index
	})
	for _, indPubKey := range indPubKeys {
		index := SEDAKeyIndex(indPubKey.Index)
		keyValidator, exists := sedaPubKeyValidators[index]
		if !exists {
			return fmt.Errorf("invalid SEDA key index %d", indPubKey.Index)
		}
		ok := keyValidator(indPubKey.PubKey)
		if !ok {
			return fmt.Errorf("invalid public key at SEDA key index %d", indPubKey.Index)
		}
	}
	return nil
}

// PubKeyToAddress converts a public key in the 65-byte uncompressed
// format into the Ethereum address format, which is defined as the
// rightmost 160 bits of Keccak hash of an ECDSA public key without
// the 0x04 prefix.
func PubKeyToEthAddress(uncompressed []byte) ([]byte, error) {
	if len(uncompressed) != 65 {
		return nil, fmt.Errorf("invalid public key length: %d", len(uncompressed))
	}
	return ethcrypto.Keccak256(uncompressed[1:])[12:], nil
}
