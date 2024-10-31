package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// generatePubKeysAndValAddrs randomly generates a given number of
// public keys encoded in uncompressed format and validator addresses.
func (s *KeeperTestSuite) generatePubKeysAndValAddrs(num int) ([][]byte, []sdk.ValAddress) {
	var pubKeys [][]byte
	var valAddrs []sdk.ValAddress
	for i := 0; i < num; i++ {
		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		pubKeys = append(pubKeys, elliptic.Marshal(privKey.PublicKey, privKey.PublicKey.X, privKey.PublicKey.Y))

		valAddrs = append(valAddrs, sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()))
	}
	return pubKeys, valAddrs
}

func (s *KeeperTestSuite) TestImportExportGenesis() {
	pubKeys, valAddrs := s.generatePubKeysAndValAddrs(10)
	genState := types.GenesisState{
		ValidatorPubKeys: []types.ValidatorPubKeys{
			{
				ValidatorAddr: valAddrs[0].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{Index: 0, PubKey: pubKeys[0]},
					{Index: 1, PubKey: pubKeys[1]},
					{Index: 2, PubKey: pubKeys[2]},
					{Index: 3, PubKey: pubKeys[3]},
				},
			},
			{
				ValidatorAddr: valAddrs[1].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{Index: 0, PubKey: pubKeys[4]},
					{Index: 2, PubKey: pubKeys[5]},
				},
			},
			{
				ValidatorAddr: valAddrs[2].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{Index: 0, PubKey: pubKeys[6]},
					{Index: 3, PubKey: pubKeys[7]},
				},
			},
			{
				ValidatorAddr: valAddrs[3].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{Index: 0, PubKey: pubKeys[8]},
				},
			},
			{
				ValidatorAddr: valAddrs[4].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{Index: 1, PubKey: pubKeys[9]},
				},
			},
		},
	}

	s.keeper.InitGenesis(s.ctx, genState)
	exportedGenState := s.keeper.ExportGenesis(s.ctx)
	s.Require().ElementsMatch(genState.ValidatorPubKeys, exportedGenState.ValidatorPubKeys)
}
