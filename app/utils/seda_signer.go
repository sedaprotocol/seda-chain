package utils

import (
	"bytes"
	"fmt"
	"sort"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

type SEDASigner interface {
	GetValAddress() sdk.ValAddress
	Sign(input []byte, index sedatypes.SEDAKeyIndex) (signature []byte, err error)
	ReloadIfMismatch(pubKeys []pubkeytypes.IndexedPubKey) error
	IsLoaded() bool
	GetPublicKeys() []pubkeytypes.IndexedPubKey
}

var _ SEDASigner = &sedaKeys{}

type sedaKeys struct {
	valAddr  sdk.ValAddress
	keys     map[sedatypes.SEDAKeyIndex]indexedPrivKey
	pubKeys  []pubkeytypes.IndexedPubKey // sorted by index
	keyPath  string
	isLoaded bool
}

// LoadSEDASigner loads the SEDA keys from a given file path and
// returns a SEDASigner interface.
func LoadSEDASigner(keyFilePath string, allowUnencrypted bool) (SEDASigner, error) {
	keys, err := loadSEDAKeys(keyFilePath, allowUnencrypted)
	if err != nil {
		keys.keyPath = keyFilePath
		keys.isLoaded = false
		return &keys, err
	}
	return &keys, nil
}

// LoadEmptySEDASigner returns an unloaded SEDASigner interface with
// only the key file path set.
func LoadEmptySEDASigner(keyFilePath string) SEDASigner {
	return &sedaKeys{
		keyPath:  keyFilePath,
		isLoaded: false,
	}
}

func loadSEDAKeys(keyFilePath string, allowUnencrypted bool) (keys sedaKeys, err error) {
	encryptionKey := ReadSEDAKeyEncryptionKeyFromEnv()
	if encryptionKey == "" && !allowUnencrypted {
		panic(fmt.Sprintf("SEDA key encryption key is not set, set the %s environment variable or allow unencrypted key file in app.toml", SEDAKeyEncryptionKeyEnvVar))
	}

	keyFile, err := loadSEDAKeyFile(keyFilePath, encryptionKey)
	if err != nil {
		return keys, err
	}

	keysMap := make(map[sedatypes.SEDAKeyIndex]indexedPrivKey)
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
func (s *sedaKeys) Sign(input []byte, index sedatypes.SEDAKeyIndex) ([]byte, error) {
	if !s.isLoaded {
		return nil, fmt.Errorf("signer is not loaded")
	}

	var signature []byte
	var err error
	switch index {
	case sedatypes.SEDAKeyIndexSecp256k1:
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

func (s *sedaKeys) GetPublicKeys() []pubkeytypes.IndexedPubKey {
	return s.pubKeys
}

// Reload reloads the signer from the key file.
func (s *sedaKeys) reload() error {
	// Reload should run from the same process as the one that loaded the signer,
	// so the check for the encryption key should already have passed if we're
	// hitting this function.
	keys, err := loadSEDAKeys(s.keyPath, true)
	if err != nil {
		s.valAddr = nil
		s.keys = nil
		s.pubKeys = nil
		s.isLoaded = false
		return err
	}

	s.valAddr = keys.valAddr
	s.keys = keys.keys
	s.pubKeys = keys.pubKeys
	s.isLoaded = true
	return nil
}
