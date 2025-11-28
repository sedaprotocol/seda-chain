package testutil

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vrf "github.com/sedaprotocol/vrf-go"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type TestAccount struct {
	name       string
	addr       sdk.AccAddress
	signingKey secp256k1.PrivKey
	fixture    *Fixture
}

func (ta *TestAccount) Name() string {
	return ta.name
}

func (ta *TestAccount) Address() string {
	return ta.addr.String()
}

func (ta *TestAccount) AccAddress() sdk.AccAddress {
	return ta.addr
}

func (ta *TestAccount) PublicKeyHex() string {
	return hex.EncodeToString(ta.signingKey.PubKey().Bytes())
}

func (ta *TestAccount) Balance() math.Int {
	return ta.fixture.BankKeeper.GetBalance(ta.fixture.Context(), ta.addr, BondDenom).Amount
}

func (ta *TestAccount) Prove(hash []byte) string {
	vrf, err := vrf.NewK256VRF().Prove(ta.signingKey.Bytes(), hash)
	require.NoError(ta.fixture.tb, err)
	return hex.EncodeToString(vrf)
}

// ExpectedPayoutUniformCase checks that the executor has received the correct amount
// of payouts based on the given numbers, assuming a uniform reporting case.
func (ta *TestAccount) CheckGasPayoutUniformCase(gasUsed uint64, gasPrice math.Int, reduced bool) (math.Int, math.Int) {
	stakerInfo, err := ta.GetStaker()
	require.NoError(ta.fixture.tb, err)

	gasUsedInt := math.NewIntFromUint64(gasUsed)
	payoutAmount := gasUsedInt.Mul(gasPrice)
	burnAmount := math.ZeroInt()

	if reduced {
		tallyConfig, err := ta.fixture.CoreKeeper.GetTallyConfig(ta.fixture.Context())
		require.NoError(ta.fixture.tb, err)

		burnAmount = tallyConfig.BurnRatio.MulInt(gasUsedInt).TruncateInt()
		payoutAmount = gasUsedInt.Sub(burnAmount).Mul(gasPrice)
	}

	require.Equal(ta.fixture.tb, payoutAmount.String(), stakerInfo.Staker.PendingWithdrawal.String())
	return payoutAmount, burnAmount
}

func (ta *TestAccount) GetDataRequestConfig() (*types.QueryDataRequestConfigResponse, error) {
	msg := &types.QueryDataRequestConfigRequest{}
	return ta.fixture.CoreQuerier.DataRequestConfig(ta.fixture.Context(), msg)
}

func HashStringHelper(input string) []byte {
	return crypto.Keccak256([]byte(input))
}

func RevealHelperFromString(input string) []byte {
	return []byte(base64.StdEncoding.EncodeToString(HashStringHelper(input)))
}

func (ta TestAccount) CreatePostDRMsg(nonce string, replicationFactor uint32) types.MsgPostDataRequest {
	execProgramID := hex.EncodeToString(HashStringHelper(nonce))
	execInputs := base64.StdEncoding.EncodeToString(HashStringHelper("exec_inputs"))
	tallyProgramID := hex.EncodeToString(HashStringHelper("tally_program"))
	tallyInputs := base64.StdEncoding.EncodeToString(HashStringHelper("tally_inputs"))

	memo := base64.StdEncoding.EncodeToString(crypto.Keccak256([]byte(ta.fixture.ChainID), []byte(nonce)))

	return types.MsgPostDataRequest{
		Sender:            ta.Address(),
		Version:           "1.0.0",
		ExecProgramID:     execProgramID,
		ExecInputs:        []byte(execInputs),
		ExecGasLimit:      types.MinExecGasLimit,
		TallyProgramID:    tallyProgramID,
		TallyInputs:       []byte(tallyInputs),
		TallyGasLimit:     types.MinTallyGasLimit,
		Memo:              []byte(memo),
		ReplicationFactor: replicationFactor,
		ConsensusFilter:   []byte{0},
		GasPrice:          types.MinGasPrice,
	}
}

func (ta *TestAccount) PostDataRequest(msg types.MsgPostDataRequest, funds *math.Int) (*types.MsgPostDataRequestResponse, error) {
	if funds != nil {
		msg.Funds = sdk.NewCoin(BondDenom, *funds)
	} else {
		msg.Funds = sdk.NewCoin(BondDenom, math.NewIntFromUint64(msg.ExecGasLimit).Add(math.NewIntFromUint64(msg.TallyGasLimit)).Mul(msg.GasPrice))
	}
	return ta.fixture.CoreMsgServer.PostDataRequest(ta.fixture.Context(), &msg)
}

func (ta *TestAccount) CreateRevealMsg(revealBody *types.RevealBody) *types.MsgReveal {
	msg := &types.MsgReveal{
		Sender:     ta.Address(),
		RevealBody: revealBody,
		PublicKey:  ta.PublicKeyHex(),
		Stderr:     []string{},
		Stdout:     []string{},
	}
	msg.Proof = ta.Prove(msg.MsgHash("", ta.fixture.ChainID))
	return msg
}

func (ta *TestAccount) CommitResult(revealMsg *types.MsgReveal) (*types.MsgCommitResponse, error) {
	msg := &types.MsgCommit{
		Sender:    ta.Address(),
		DrID:      revealMsg.RevealBody.DrID,
		Commit:    hex.EncodeToString(revealMsg.RevealHash()),
		PublicKey: ta.PublicKeyHex(),
	}
	//nolint:gosec // G115: Block height is never negative.
	msg.Proof = ta.Prove(msg.MsgHash("", ta.fixture.ChainID, int64(revealMsg.RevealBody.DrBlockHeight)))

	ta.fixture.SetTx(100_000, ta.AccAddress(), msg)
	res, err := ta.fixture.CoreMsgServer.Commit(ta.fixture.Context(), msg)
	ta.fixture.SetInfiniteGasMeter()
	return res, err
}

func (ta *TestAccount) RevealResult(msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	ta.fixture.SetTx(100_000, ta.AccAddress(), msg)
	res, err := ta.fixture.CoreMsgServer.Reveal(ta.fixture.Context(), msg)
	ta.fixture.SetInfiniteGasMeter()
	return res, err
}

func (ta *TestAccount) GetDataRequestsByStatus(status types.DataRequestStatus, limit uint64, lastSeenIndex *[]string) (*types.QueryDataRequestsByStatusResponse, error) {
	msg := &types.QueryDataRequestsByStatusRequest{
		Status: status,
		Limit:  limit,
	}
	if lastSeenIndex != nil {
		msg.LastSeenIndex = *lastSeenIndex
	}

	return ta.fixture.CoreQuerier.DataRequestsByStatus(ta.fixture.Context(), msg)
}

func (ta *TestAccount) GetDataRequestStatuses(drIDs []string) (*types.QueryDataRequestStatusesResponse, error) {
	msg := &types.QueryDataRequestStatusesRequest{
		DataRequestIds: drIDs,
	}
	return ta.fixture.CoreQuerier.DataRequestStatuses(ta.fixture.Context(), msg)
}

func (ta *TestAccount) GetDataRequest(drID string) (*types.QueryDataRequestResponse, error) {
	msg := &types.QueryDataRequestRequest{
		DrId: drID,
	}
	return ta.fixture.CoreQuerier.DataRequest(ta.fixture.Context(), msg)
}
