package utils_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
)

type SEDAKeysTestSuite struct {
	suite.Suite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SEDAKeysTestSuite))
}

func (s *SEDAKeysTestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	config.Seal()
}

func (s *SEDAKeysTestSuite) TestSEDAKeyEncryptionDecryption() {
	valAddr, err := sdk.ValAddressFromBech32("sedavaloper12rype4zl8wxcgqwl237fll6hvufkgcj8act8xw")
	require.NoError(s.T(), err)

	encryptionKey, err := utils.GenerateSEDAKeyEncryptionKey()
	require.NoError(s.T(), err)

	tempDir := s.T().TempDir()
	keyfilePath := filepath.Join(tempDir, "seda_keys.json")
	generatedKeys, err := utils.GenerateSEDAKeys(valAddr, keyfilePath, encryptionKey, false)

	s.Require().NoError(err)
	s.Require().Equal(sedatypes.SEDAKeyIndex(generatedKeys[0].Index), sedatypes.SEDAKeyIndexSecp256k1)
	s.Require().NotEmpty(generatedKeys[0].PubKey, "public key should not be empty")

	invalidKey, err := utils.GenerateSEDAKeyEncryptionKey()
	s.Require().NoError(err)

	_, err = utils.LoadSEDAPubKeys(keyfilePath, invalidKey)
	s.Require().ErrorContains(err, "cipher: message authentication failed")

	loadedKeys, err := utils.LoadSEDAPubKeys(keyfilePath, encryptionKey)
	s.Require().NoError(err)
	s.Require().Equal(generatedKeys[0].PubKey, loadedKeys[0].PubKey)
}

func (s *SEDAKeysTestSuite) TestSEDAKeyDecryptionExistingFile() {
	tempDir := s.T().TempDir()
	err := os.WriteFile(filepath.Join(tempDir, "seda_keys.json"), []byte("kNYhCAjfN9BhJ46iYzJWCUXn9efOAGf30D81UjF5tRlRtdiziW1zGVK+6ehxeJXKcPAmWjQkTxAKcJv7ozAA0xdleR4yO6HakROtFRXlOBy3K9Fv6rkDfCmbIUUjOH9oGP2F5+ldKeE5030MOdNORWUKW7fIlnKUyBWTZfLSmsKi+iCaIyZ/bFh2+NDiESPHAYl+X8t+SKKy6MgAwarrW9W1/6enNLoVmF8dAJ1dhxeKyXF/aXWKR7HaMRwe7V1NjfnaFcI09CeibpWud9rYKhbjV3K0/RdBobjPTIHAnLd5erh/3eVo9RGm8bC8a97obKm68lDernSN9HvjoTO3QlvI0k7cVDAhiuphS4qlgjOVW+eWm+S5dlD2gpCExcmrqxbggLOtjoZbQyrKhQFmfn5UGonoDTSbwtbZZtvY1N48AVT4eueReBWumcipO0ViWnkxLNIJ8vFA"), 0o600)
	s.Require().NoError(err)

	_, err = utils.LoadSEDAPubKeys(filepath.Join(tempDir, "seda_keys.json"), "xmp1EDn7ndgZIdgwupJ9yfDWlSssubKpgo2ZHqjx+4w=")
	s.Require().ErrorContains(err, "cipher: message authentication failed")

	keys, err := utils.LoadSEDAPubKeys(filepath.Join(tempDir, "seda_keys.json"), "La1PSNwUBZXEoIQ1CM0VF+kRr9vqforxE97afYdTF+c=")
	s.Require().NoError(err)
	s.Require().Equal(hex.EncodeToString(keys[0].PubKey), "04be41e55492d9d823c435b6b6801413223b31fdfa0318d2dea51e1886215e8664e234c34afa7af32ec02a1d0289ce656bab3ed106646836c9d26ce35968b2ff68")
}

func (s *SEDAKeysTestSuite) TestSEDAKeyWithoutEncryption() {
	valAddr, err := sdk.ValAddressFromBech32("sedavaloper12rype4zl8wxcgqwl237fll6hvufkgcj8act8xw")
	require.NoError(s.T(), err)

	tempDir := s.T().TempDir()
	keyfilePath := filepath.Join(tempDir, "seda_keys.json")
	generatedKeys, err := utils.GenerateSEDAKeys(valAddr, keyfilePath, "", false)

	s.Require().NoError(err)
	s.Require().Equal(sedatypes.SEDAKeyIndex(generatedKeys[0].Index), sedatypes.SEDAKeyIndexSecp256k1)
	s.Require().NotEmpty(generatedKeys[0].PubKey, "public key should not be empty")

	keys, err := os.ReadFile(keyfilePath)
	s.Require().NoError(err)

	// We only verify the top level of the JSON schema, this should be enough
	// to ensure that the file was not encrypted.
	type jsonSchema struct {
		ValidatorAddr sdk.ValAddress `json:"validator_addr"`
		Keys          []interface{}  `json:"keys"`
	}

	var sedaKeyFile jsonSchema
	s.Require().NoError(json.Unmarshal(keys, &sedaKeyFile))
	s.Require().Equal(sedaKeyFile.ValidatorAddr, valAddr)
	s.Require().Equal(len(sedaKeyFile.Keys), 1)

	// Test that the file can be loaded without encryption.
	loadedKeys, err := utils.LoadSEDAPubKeys(keyfilePath, "")
	s.Require().NoError(err)
	s.Require().Equal(generatedKeys[0].PubKey, loadedKeys[0].PubKey)
}
