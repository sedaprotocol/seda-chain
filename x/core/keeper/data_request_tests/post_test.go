package datarequesttests

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestPostWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// try to get dr that doesn't exist yet
	dr := bob.CreatePostDRMsg("1", 1)
	drID, err := dr.ComputeDataRequestID()
	require.NoError(t, err)

	_, err = bob.GetDataRequest(drID)
	require.ErrorContains(t, err, "not found")

	// Bob posts a data request
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// Dr can now be found
	drFound, err := bob.GetDataRequest(drID)
	require.NoError(t, err)
	require.Equal(t, postDrResult.DrID, drFound.DataRequest.ID)

	// check the escrow for the DR is correct
	expectedEscrow := math.NewIntFromUint64(dr.ExecGasLimit).Add(math.NewIntFromUint64(dr.TallyGasLimit)).Mul(dr.GasPrice)
	require.Equal(t, expectedEscrow, drFound.DataRequest.Escrow)
}

func TestCannotDoublePost(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CreatePostDRMsg("1", 1)
	_, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// Bob tries to post the same data request again
	_, err = bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "data request already exists")
}

func TestFailsIfNotEnoughFunds(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 1)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CreatePostDRMsg("1", 1)
	dr.GasPrice = math.NewInt(100_000_000_000)
	insufficientFunds := math.NewInt(1)
	_, err := bob.PostDataRequest(dr, &insufficientFunds)
	require.ErrorContains(t, err, "insufficient funds")
}

func TestWithMaxGasLimits(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	int64Max := int64(^uint64(0) >> 1)
	bob := f.CreateTestAccount("bob", int64Max)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with max gas limit
	dr := bob.CreatePostDRMsg("1", 1)
	uint64Max := ^uint64(0)
	dr.ExecGasLimit = uint64Max
	dr.TallyGasLimit = uint64Max

	funds := (math.NewIntFromUint64(uint64Max).Add(math.NewIntFromUint64(uint64Max))).Mul(types.MinGasPrice)
	_, err := bob.PostDataRequest(dr, &funds)
	require.NoError(t, err)
}

func TestFailsIfReplicationFactorTooHigh(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with too high replication factor
	dr := bob.CreatePostDRMsg("1", 2)
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "replication factor is too high")
}

func TestFailsIfReplicationFactorIsZero(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	dr := bob.CreatePostDRMsg("1", 0)
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "replication factor cannot be zero")
}

func TestFailsIfMinGasPriceIsNotMet(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with too low gas price
	dr := bob.CreatePostDRMsg("1", 1)
	dr.GasPrice = types.MinGasPrice.Sub(math.NewInt(1))
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "gas price is too low")
}

func TestFailsIfMinGasExecLimitIsNotMet(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with too low exec gas limit
	dr := bob.CreatePostDRMsg("1", 1)
	dr.ExecGasLimit = types.MinExecGasLimit - 1
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "exec gas limit is too low")
}

func TestFailsIfMinGasTallyLimitIsNotMet(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with too low tally gas limit
	dr := bob.CreatePostDRMsg("1", 1)
	dr.TallyGasLimit = types.MinTallyGasLimit - 1
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "tally gas limit is too low")
}

func TestFailsIfInvalidExecProgramID(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with invalid exec program ID (not hex)
	dr := bob.CreatePostDRMsg("1", 1)
	dr.ExecProgramID = "invalid_hex"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "exec program ID is not a valid hex string")
}

func TestFailsIfInvalidTallyProgramID(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with invalid tally program ID (not hex)
	dr := bob.CreatePostDRMsg("1", 1)
	dr.TallyProgramID = "invalid_hex"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "tally program ID is not a valid hex string")
}

func TestFailsIfInvalidLengthExecProgramID(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with invalid exec program ID (wrong length)
	dr := bob.CreatePostDRMsg("1", 1)
	dr.ExecProgramID = "deadbeef"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "hex-encoded exec program ID is not 64 characters long")
}

func TestFailsIfInvalidLengthTallyProgramID(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with invalid tally program ID (wrong length)
	dr := bob.CreatePostDRMsg("1", 1)
	dr.TallyProgramID = "deadbeef"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "hex-encoded tally program ID is not 64 characters long")
}

func TestFailsIfExecInputsTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big exec inputs
	dr := bob.CreatePostDRMsg("1", 1)
	execInputs := make([]byte, drConfigResp.DataRequestConfig.ExecInputLimitInBytes+1)
	dr.ExecInputs = execInputs
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "exec input limit exceeded")
}

func TestFailsIfTallyInputsTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big tally inputs
	dr := bob.CreatePostDRMsg("1", 1)
	tallyInputs := make([]byte, drConfigResp.DataRequestConfig.TallyInputLimitInBytes+1)
	dr.TallyInputs = tallyInputs
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "tally input limit exceeded")
}

func TestFailsIfConsensusFilterTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big consensus filter
	dr := bob.CreatePostDRMsg("1", 1)
	consensusFilter := make([]byte, drConfigResp.DataRequestConfig.ConsensusFilterLimitInBytes+1)
	dr.ConsensusFilter = consensusFilter
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "consensus filter limit exceeded")
}

func TestFailsIfMemoTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big memo
	dr := bob.CreatePostDRMsg("1", 1)
	memo := make([]byte, drConfigResp.DataRequestConfig.MemoLimitInBytes+1)
	dr.Memo = memo
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "memo limit exceeded")
}

func TestFailsIfPaybackAddressTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big payback address
	dr := bob.CreatePostDRMsg("1", 1)
	paybackAddress := make([]byte, drConfigResp.DataRequestConfig.PaybackAddressLimitInBytes+1)
	dr.PaybackAddress = paybackAddress
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "payback address limit exceeded")
}

func TestFailsIfSedaPayloadTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)

	// Bob posts a data request with too big seda payload
	dr := bob.CreatePostDRMsg("1", 1)
	sedaPayload := make([]byte, drConfigResp.DataRequestConfig.SEDAPayloadLimitInBytes+1)
	dr.SEDAPayload = sedaPayload
	_, err = bob.PostDataRequest(dr, nil)
	t.Log(err)
	require.ErrorContains(t, err, "SEDA payload limit exceeded")
}

func TestFailsIfVersionHasPre(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with version containing "pre"
	dr := bob.CreatePostDRMsg("1", 1)
	dr.Version = "1.0.0-pre"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "invalid data request version")
}

func TestFailsIfVersionHasBuildMetadata(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with version containing build metadata
	dr := bob.CreatePostDRMsg("1", 1)
	dr.Version = "1.0.0+build.1"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "invalid data request version")
}

func TestFailsIfVersionHasBothPreAndBuildMetadata(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request with version containing both pre-release and build metadata
	dr := bob.CreatePostDRMsg("1", 1)
	dr.Version = "1.0.0-pre+build.1"
	_, err := bob.PostDataRequest(dr, nil)
	require.ErrorContains(t, err, "invalid data request version")
}
