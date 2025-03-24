package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cast"

	cmtos "github.com/cometbft/cometbft/libs/os"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const (
	// FlagAllowUnencryptedSedaKeys is a flag that allows unencrypted SEDA keys.
	FlagAllowUnencryptedSedaKeys = "allow-unencrypted-seda-keys"
	// EnvAllowUnencryptedSedaKeys is an environment variable that allows unencrypted SEDA keys.
	EnvAllowUnencryptedSedaKeys = "SEDA_ALLOW_UNENCRYPTED_KEYS"
	// SEDAKeyEncryptionKeyEnvVar is the environment variable that should contain the SEDA key encryption key.
	SEDAKeyEncryptionKeyEnvVar = "SEDA_KEYS_ENCRYPTION_KEY"
)

func ShouldAllowUnencryptedSedaKeys(appOpts servertypes.AppOptions) bool {
	allowUnencryptedFlag := cast.ToBool(appOpts.Get(FlagAllowUnencryptedSedaKeys))
	_, allowUnencryptedInEnv := os.LookupEnv(EnvAllowUnencryptedSedaKeys)

	return allowUnencryptedFlag || allowUnencryptedInEnv
}

// ReadSEDAKeyEncryptionKeyFromEnv reads the SEDA key encryption key from
// the environment variable. Returns an empty string if the environment
// variable is not set.
func ReadSEDAKeyEncryptionKeyFromEnv() string {
	return os.Getenv(SEDAKeyEncryptionKeyEnvVar)
}

func GenerateSEDAKeyEncryptionKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

type privKeyGenerator func() *ecdsa.PrivateKey

// sedaKeyGenerators maps the SEDA key index to the corresponding
// private key generator.
var sedaKeyGenerators = map[sedatypes.SEDAKeyIndex]privKeyGenerator{
	sedatypes.SEDAKeyIndexSecp256k1: func() *ecdsa.PrivateKey {
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
var sedaPubKeyValidators = map[sedatypes.SEDAKeyIndex]pubKeyValidator{
	sedatypes.SEDAKeyIndexSecp256k1: func(pub []byte) bool {
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
	Index   sedatypes.SEDAKeyIndex `json:"index"`
	PrivKey *ecdsa.PrivateKey      `json:"priv_key"`
	PubKey  []byte                 `json:"pub_key"`
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

// saveSEDAKeyFile saves a given list of indexedPrivKey in the directory
// at dirPath. When encryptionKey is not empty, the file is encrypted
// using the provided key and stored as base64 encoded.
func saveSEDAKeyFile(keys []indexedPrivKey, valAddr sdk.ValAddress, dirPath string, encryptionKey string, forceKeyFile bool) error {
	savePath := filepath.Join(dirPath, SEDAKeyFileName)
	if SEDAKeyFileExists(dirPath) && !forceKeyFile {
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

	if encryptionKey != "" {
		encryptedData, err := encryptBytes(jsonBytes, encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt SEDA keys: %v", err)
		}
		jsonBytes = []byte(base64.StdEncoding.EncodeToString(encryptedData))
	}

	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write SEDA key file: %v", err)
	}
	return nil
}

func SEDAKeyFileExists(dirPath string) bool {
	return cmtos.FileExists(filepath.Join(dirPath, SEDAKeyFileName))
}

// loadSEDAKeyFile loads the SEDA key file from the given path. When
// encryptionKey is not empty, the file is processed as base64 encoded
// and then decrypted using the provided key.
func loadSEDAKeyFile(loadPath string, encryptionKey string) (sedaKeyFile, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return sedaKeyFile{}, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}

	if encryptionKey != "" {
		decodedBytes, err := base64.StdEncoding.DecodeString(string(keysJSONBytes))
		if err != nil {
			return sedaKeyFile{}, fmt.Errorf("failed to base64 decode SEDA keys: %v", err)
		}
		decryptedData, err := decryptBytes(decodedBytes, encryptionKey)
		if err != nil {
			return sedaKeyFile{}, fmt.Errorf("failed to decrypt SEDA keys: %v", err)
		}
		keysJSONBytes = decryptedData
	}

	var keyFile sedaKeyFile
	err = json.Unmarshal(keysJSONBytes, &keyFile)
	if err != nil {
		return sedaKeyFile{}, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}
	return keyFile, nil
}

// LoadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs. When encryptionKey is not
// empty, the file is processed as base64 encoded and then decrypted
// using the provided key.
func LoadSEDAPubKeys(loadPath string, encryptionKey string) ([]pubkeytypes.IndexedPubKey, error) {
	keyFile, err := loadSEDAKeyFile(loadPath, encryptionKey)
	if err != nil {
		return nil, err
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
// in the directory given by dirPath. When encryptionKey is not empty,
// the file is encrypted using the provided key and stored as base64
// encoded. If forceKeyFile is true, the key file is overwritten if it
// already exists.
func GenerateSEDAKeys(valAddr sdk.ValAddress, dirPath string, encryptionKey string, forceKeyFile bool) ([]pubkeytypes.IndexedPubKey, error) {
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
	err := saveSEDAKeyFile(privKeys, valAddr, dirPath, encryptionKey, forceKeyFile)
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
		index := sedatypes.SEDAKeyIndex(indPubKey.Index)
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

func encryptBytes(data []byte, key string) ([]byte, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	aes, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptBytes(data []byte, key string) ([]byte, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	aes, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	decryptedData, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}
