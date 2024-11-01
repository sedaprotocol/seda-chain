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

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// SEDAKeyFileName defines the SEDA key file name.
const SEDAKeyFileName = "seda_keys.json"

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

// sedaKeyGenerators maps the key index to the corresponding private
// key generator.
var sedaKeyGenerators = map[SEDAKeyIndex]privKeyGenerator{
	SEDAKeyIndexSecp256k1: func() *ecdsa.PrivateKey {
		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		return privKey
	},
}

var sedaKeyValidators = map[SEDAKeyIndex]pubKeyValidator{
	SEDAKeyIndexSecp256k1: func(pub []byte) bool {
		_, err := ethcrypto.UnmarshalPubkey(pub)
		return err == nil
	},
}

type privKeyGenerator func() *ecdsa.PrivateKey

type pubKeyValidator func([]byte) bool

type indexedPrivKey struct {
	Index   SEDAKeyIndex      `json:"index"`
	PrivKey *ecdsa.PrivateKey `json:"priv_key"`
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

// GenerateSEDAKeys generates SEDA keys given a list of private key
// generators, saves them to the SEDA key file, and returns the resulting
// index-public key pairs. Index is assigned incrementally in the order
// of the given private key generators. The key file is stored in the
// directory given by dirPath.
func GenerateSEDAKeys(dirPath string) ([]pubkeytypes.IndexedPubKey, error) {
	keys := make([]indexedPrivKey, len(sedaKeyGenerators))
	result := make([]pubkeytypes.IndexedPubKey, len(sedaKeyGenerators))
	for i, generator := range sedaKeyGenerators {
		keys[i] = indexedPrivKey{
			Index:   i,
			PrivKey: generator(),
		}
		pubKey := keys[i].PrivKey.PublicKey
		pubKeyBytes := ethcrypto.FromECDSAPub(&pubKey)
		result[i] = pubkeytypes.IndexedPubKey{
			Index:  uint32(i),
			PubKey: pubKeyBytes,
		}
	}

	// The key file is placed in the same directory as the validator key file.
	err := saveSEDAKeys(keys, dirPath)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ValidateSEDAKeys ensures that the provided indexed public keys
// conform to SEDA keys specifications. It first sorts the provided
// slice for deterministic results.
func ValidateSEDAKeys(indPubKeys []pubkeytypes.IndexedPubKey) error {
	if len(sedaKeyValidators) != len(indPubKeys) {
		return fmt.Errorf("invalid number of SEDA keys")
	}
	sort.Slice(indPubKeys, func(i, j int) bool {
		return indPubKeys[i].Index < indPubKeys[j].Index
	})
	for _, indPubKey := range indPubKeys {
		index := SEDAKeyIndex(indPubKey.Index)
		keyValidator, exists := sedaKeyValidators[index]
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

// LoadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs.
func LoadSEDAPubKeys(loadPath string) ([]pubkeytypes.IndexedPubKey, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []indexedPrivKey
	err = json.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	result := make([]pubkeytypes.IndexedPubKey, len(keys))
	for i, key := range keys {
		pubKey := key.PrivKey.PublicKey
		pubKeyBytes := ethcrypto.FromECDSAPub(&pubKey)
		result[i] = pubkeytypes.IndexedPubKey{
			Index:  uint32(key.Index),
			PubKey: pubKeyBytes,
		}
	}
	return result, nil
}

// saveSEDAKeys saves a given list of IndexedPrivKey in the directory
// at dirPath.
func saveSEDAKeys(keys []indexedPrivKey, dirPath string) error {
	savePath := filepath.Join(dirPath, SEDAKeyFileName)
	if cmtos.FileExists(savePath) {
		return fmt.Errorf("SEDA key file already exists at %s", savePath)
	}
	err := cmtos.EnsureDir(filepath.Dir(savePath), 0o700)
	if err != nil {
		return err
	}
	jsonBytes, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SEDA keys: %v", err)
	}
	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write SEDA key file: %v", err)
	}
	return nil
}

type SEDASigner interface {
	Sign(input []byte, index SEDAKeyIndex) (signature []byte, err error)
}

var _ SEDASigner = &sedaKeys{}

type sedaKeys struct {
	keys []indexedPrivKey
}

// LoadSEDASigner loads the SEDA keys from a given file and returns
// a SEDASigner interface.
func LoadSEDASigner(loadPath string) (SEDASigner, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []indexedPrivKey
	err = json.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}
	return &sedaKeys{keys: keys}, nil
}

// Sign signs a 32-byte digest with the key at the given index.
func (s *sedaKeys) Sign(input []byte, index SEDAKeyIndex) ([]byte, error) {
	var signature []byte
	var err error
	switch index {
	case SEDAKeyIndexSecp256k1:
		signature, err = ethcrypto.Sign(input, s.keys[index].PrivKey)
	default:
		err = fmt.Errorf("invalid SEDA key index %d", index)
	}
	if err != nil {
		return nil, err
	}
	return signature, nil
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
