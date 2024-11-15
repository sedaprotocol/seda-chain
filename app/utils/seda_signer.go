package utils

import (
	"bytes"
	"fmt"
	"sort"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type SEDASigner interface {
	GetValAddress() sdk.ValAddress
	Sign(input []byte, index SEDAKeyIndex) (signature []byte, err error)
	ReloadIfMismatch(pubKeys []pubkeytypes.IndexedPubKey) error
	IsLoaded() bool
}

var _ SEDASigner = &sedaKeys{}

type sedaKeys struct {
	valAddr  sdk.ValAddress
	keys     map[SEDAKeyIndex]indexedPrivKey
	pubKeys  []pubkeytypes.IndexedPubKey // sorted by index
	keyPath  string
	isLoaded bool
}

// LoadSEDASigner loads the SEDA keys from a given file path and
// returns a SEDASigner interface.
func LoadSEDASigner(keyFilePath string) (SEDASigner, error) {
	keys, err := loadSEDAKeys(keyFilePath)
	if err != nil {
		keys.keyPath = keyFilePath
		keys.isLoaded = false
		return &keys, err
	}
	return &keys, nil
}

func loadSEDAKeys(keyFilePath string) (keys sedaKeys, err error) {
	keyFile, err := loadSEDAKeyFile(keyFilePath)
	if err != nil {
		return keys, err
	}

	keysMap := make(map[SEDAKeyIndex]indexedPrivKey)
	indPubKeys := make([]pubkeytypes.IndexedPubKey, len(keyFile.Keys))
	for _, key := range keyFile.Keys {
		keysMap[key.Index] = key
		indPubKeys[key.Index] = pubkeytypes.IndexedPubKey{
			Index:  uint32(key.Index),
			PubKey: key.PubKey,
		}
	}
	sort.Slice(indPubKeys, func(i, j int) bool {
		return indPubKeys[i].Index < indPubKeys[j].Index
	})

	keys.valAddr = keyFile.ValidatorAddr
	keys.keys = keysMap
	keys.pubKeys = indPubKeys
	keys.keyPath = keyFilePath
	keys.isLoaded = true
	return keys, nil
}

// GetConsAddress returns the signer's validator address.
func (s *sedaKeys) GetValAddress() sdk.ValAddress {
	return s.valAddr
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

// ReloadIfMismatch reloads the signer if the given indexed public keys
// do not match the currently loaded ones. If no indexed public keys are
// given, the signer is reloaded.
func (s *sedaKeys) ReloadIfMismatch(pubKeys []pubkeytypes.IndexedPubKey) error {
	if len(pubKeys) == 0 {
		return s.reload()
	}
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

// IsLoaded returns true if the signer is loaded and ready for signing.
func (s *sedaKeys) IsLoaded() bool {
	return s.isLoaded
}

// Reload reloads the signer from the key file.
func (s *sedaKeys) reload() error {
	keys, err := loadSEDAKeys(s.keyPath)
	if err != nil {
		return err
	}

	s.valAddr = keys.valAddr
	s.keys = keys.keys
	s.pubKeys = keys.pubKeys
	s.isLoaded = true
	return nil
}
