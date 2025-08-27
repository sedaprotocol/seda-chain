package keeper_test

import (
	"encoding/hex"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestOracleProgram() {
	s.SetupTest()
	s.ApplyDefaultMockExpectations()
	wasm := testwasms.HelloWorldWasm()
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreOracleProgram{
		Sender:     s.authority,
		Wasm:       compWasm,
		StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(wasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
	}
	storedWasm, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryOracleProgramRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.OracleProgram(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.OracleProgram.Hash))
}

func (s *KeeperTestSuite) TestOraclePrograms() {
	s.SetupTest()
	s.ApplyDefaultMockExpectations()

	wasm := testwasms.SampleTallyWasm()
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreOracleProgram{
		Sender:     s.authority,
		Wasm:       compWasm,
		StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(wasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
	}
	storedWasm, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
	s.Require().NoError(err)

	wasm2 := testwasms.SampleTallyWasm2()
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreOracleProgram{
		Sender:     s.authority,
		Wasm:       compWasm2,
		StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(wasm2))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
	}
	storedWasm2, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryOracleProgramsRequest{}
	res, err := s.queryClient.OraclePrograms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Contains(res.List, storedWasm.Hash)
	s.Require().Contains(res.List, storedWasm2.Hash)
}
