package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	pvm "github.com/cometbft/cometbft/privval"

	sdk "github.com/cosmos/cosmos-sdk/types"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type SEDASigner interface {
	GetConsAddress() sdk.ConsAddress
	Sign(input []byte, index SEDAKeyIndex) (signature []byte, err error)
	ReloadIfMismatch(pubKeys []pubkeytypes.IndexedPubKey) error
}

var _ SEDASigner = &sedaKeys{}

type sedaKeys struct {
	consAddr sdk.ConsAddress
	keys     map[SEDAKeyIndex]indexedPrivKey
	pubKeys  []pubkeytypes.IndexedPubKey // sorted by index
	keyPath  string
}

// LoadSEDASigner loads the SEDA keys from a given file and returns
// a SEDASigner interface.
func LoadSEDASigner(pvKeyFilePath string) (SEDASigner, error) {
	// TODO What if there is a rotation?
	// TODO Can we safely assume that the file will be loaded?
	_, err := os.ReadFile(pvKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private validator key from %v: %v", pvKeyFilePath, err)
	}
	privValidator := pvm.LoadFilePVEmptyState(pvKeyFilePath, "")
	consAddr := (sdk.ConsAddress)(privValidator.GetAddress())

	loadPath := filepath.Join(filepath.Dir(pvKeyFilePath), SEDAKeyFileName)
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []indexedPrivKey
	err = json.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	keysMap := make(map[SEDAKeyIndex]indexedPrivKey)
	indPubKeys := make([]pubkeytypes.IndexedPubKey, len(keys))
	for _, key := range keys {
		keysMap[key.Index] = key
		indPubKeys[key.Index] = pubkeytypes.IndexedPubKey{
			Index:  uint32(key.Index),
			PubKey: key.PubKey,
		}
	}
	sort.Slice(indPubKeys, func(i, j int) bool {
		return indPubKeys[i].Index < indPubKeys[j].Index
	})

	return &sedaKeys{
		consAddr: consAddr,
		keys:     keysMap,
		pubKeys:  indPubKeys,
		keyPath:  loadPath,
	}, nil
}

// GetConsAddress returns the signer's consensus address.
func (s *sedaKeys) GetConsAddress() sdk.ConsAddress {
	return s.consAddr
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

// ReloadIfMismatch compares the given indexed public keys to the
// currently loaded public keys. If there is any mismatch, the signer
// is reloaded.
func (s *sedaKeys) ReloadIfMismatch(pubKeys []pubkeytypes.IndexedPubKey) error {
	for _, pubKey := range s.pubKeys {
		found := false
		for _, pk := range pubKeys {
			if pk.Index == pubKey.Index {
				if !bytes.Equal(pk.PubKey, pubKey.PubKey) {
					return s.reload()
				}
				found = true
			}
		}
		if !found {
			return s.reload()
		}
	}
	return nil
}

// Reload reloads the signer from the key file.
func (s *sedaKeys) reload() error {
	// TODO merge with LoadSEDASigner??
	keysJSONBytes, err := os.ReadFile(s.keyPath)
	if err != nil {
		return fmt.Errorf("failed to read SEDA keys from %v: %v", s.keyPath, err)
	}
	var keys []indexedPrivKey
	err = json.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", s.keyPath, err)
	}
	keysMap := make(map[SEDAKeyIndex]indexedPrivKey)
	for _, key := range keys {
		keysMap[key.Index] = key
	}
	s.keys = keysMap
	return nil
}
