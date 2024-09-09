package keeper_test

import (
	"github.com/cometbft/cometbft/crypto/secp256k1"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// generatePubKeysAndValAddrs randomly generates a given number of
// public keys encoded in codectypes.Any type and their validator
// addresses.
func (s *KeeperTestSuite) generatePubKeysAndValAddrs(num int) ([]*codectypes.Any, []sdk.ValAddress) {
	var pkAnys []*codectypes.Any
	var valAddrs []sdk.ValAddress
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(privKey.PubKey())
		s.Require().NoError(err)

		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		s.Require().NoError(err)
		pkAnys = append(pkAnys, pkAny)

		valAddrs = append(valAddrs, sdk.ValAddress(privKey.PubKey().Address()))
	}
	return pkAnys, valAddrs
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
