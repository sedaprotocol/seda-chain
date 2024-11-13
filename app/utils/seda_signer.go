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
}

var _ SEDASigner = &sedaKeys{}

type sedaKeys struct {
	valAddr sdk.ValAddress
	keys    map[SEDAKeyIndex]indexedPrivKey
	pubKeys []pubkeytypes.IndexedPubKey // sorted by index
	keyPath string
}

// LoadSEDASigner loads the SEDA keys from a given file path and
// returns a SEDASigner interface.
func LoadSEDASigner(keyFilePath string) (SEDASigner, error) {
	sedaKeys, err := loadSEDAKeys(keyFilePath)
	if err != nil {
		return nil, err
	}
	return sedaKeys, nil
}

func loadSEDAKeys(keyFilePath string) (*sedaKeys, error) {
	keyFile, err := loadSEDAKeyFile(keyFilePath)
	if err != nil {
		return nil, err
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

	return &sedaKeys{
		valAddr: keyFile.ValidatorAddr,
		keys:    keysMap,
		pubKeys: indPubKeys,
		keyPath: keyFilePath,
	}, nil
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
	keyFile, err := loadSEDAKeyFile(s.keyPath)
	if err != nil {
		return err
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

	s.keys = keysMap
	s.valAddr = keyFile.ValidatorAddr
	s.pubKeys = indPubKeys
	return nil
}
