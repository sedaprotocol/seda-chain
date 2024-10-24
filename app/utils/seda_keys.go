package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec/v2"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

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
	SEDAKeyIndexSecp256k1: func() cmtcrypto.PrivKey { return secp256k1.GenPrivKey() },
}

type (
	privKeyGenerator func() cmtcrypto.PrivKey

	indexedPrivKey struct {
		Index   SEDAKeyIndex      `json:"index"`
		PrivKey cmtcrypto.PrivKey `json:"priv_key"`
	}
)

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

		// Convert to SDK type for app-level use.
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(keys[i].PrivKey.PubKey())
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, err
		}
		result[i] = pubkeytypes.IndexedPubKey{
			Index:  uint32(i),
			PubKey: pkAny,
		}
	}

	// The key file is placed in the same directory as the validator key file.
	err := saveSEDAKeys(keys, dirPath)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LoadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs.
func LoadSEDAPubKeys(loadPath string) ([]pubkeytypes.IndexedPubKey, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []indexedPrivKey
	err = cmtjson.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	result := make([]pubkeytypes.IndexedPubKey, len(keys))
	for i, key := range keys {
		// Convert to SDK type for app-level use.
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(key.PrivKey.PubKey())
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, err
		}
		result[i] = pubkeytypes.IndexedPubKey{
			Index:  uint32(key.Index),
			PubKey: pkAny,
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
	jsonBytes, err := cmtjson.MarshalIndent(keys, "", "  ")
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
	err = cmtjson.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}
	return &sedaKeys{keys: keys}, nil
}

func (s *sedaKeys) Sign(input []byte, index SEDAKeyIndex) ([]byte, error) {
	signature, err := s.keys[index].PrivKey.Sign(input)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// PubKeyToAddress converts a public key in the 33-byte compressed
// format into the Ethereum address format, which is defined as the
// rightmost 160 bits of a Keccak hash of an ECDSA public key.
func PubKeyToAddress(pubkey []byte) ([]byte, error) {
	if len(pubkey) != 33 {
		return nil, fmt.Errorf("invalid compressed public key %x", pubkey)
	}
	key, err := btcec.ParsePubKey(pubkey)
	if err != nil {
		return nil, err
	}
	// 64-byte format: x-coordinate | y-coordinate
	uncompressed := append(key.X().Bytes(), key.Y().Bytes()...)
	return ethcrypto.Keccak256(uncompressed)[12:], nil
}
