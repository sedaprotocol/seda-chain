package utils

import (
	"fmt"
	"os"
	"path/filepath"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const SEDAKeyFileName = "seda_keys.json"

var sedaKeyGenerators = []privKeyGenerator{
	func() cmtcrypto.PrivKey { return secp256k1.GenPrivKey() }, // index 0 - secp256k1
}

type (
	privKeyGenerator func() cmtcrypto.PrivKey

	indexedPrivKey struct {
		Index   uint32            `json:"index"`
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
			Index:   uint32(i),
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
			Index:  key.Index,
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

type sedaSigner interface {
	Sign(input []byte, index uint32) (signature []byte, err error)
}

var _ sedaSigner = &sedaKeys{}

type sedaKeys struct {
	keys []indexedPrivKey
}

// LoadSEDASigner loads the SEDA keys from the given file and returns
// a sedaKeys object.
func LoadSEDASigner(loadPath string) (*sedaKeys, error) {
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

func (s *sedaKeys) Sign(input []byte, index uint32) ([]byte, error) {
	signature, err := s.keys[index].PrivKey.Sign(input)
	if err != nil {
		return nil, err
	}
	return signature, nil
}
