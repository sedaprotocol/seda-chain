package abci_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/sha3"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtsecp256k1 "github.com/cometbft/cometbft/crypto/secp256k1"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	"github.com/sedaprotocol/seda-chain/app/abci"
	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

var (
	chainID = "seda-abci-test"
)

type testValidator struct {
	consAddr    sdk.ConsAddress
	valAddr     sdk.ValAddress
	tmPk        cmtprotocrypto.PublicKey
	privKey     cmtsecp256k1.PrivKey
	isNew       bool
	signer      utils.SEDASigner
	handlers    *abci.Handlers
	ethAddr     []byte
	sedaPubKeys []pubkeytypes.IndexedPubKey
	voteExt     []byte
}

func newTestValidator() testValidator {
	privkey := cmtsecp256k1.GenPrivKey()
	pubkey := privkey.PubKey()
	tmPk := cmtprotocrypto.PublicKey{
		Sum: &cmtprotocrypto.PublicKey_Secp256K1{
			Secp256K1: pubkey.Bytes(),
		},
	}
	return testValidator{
		consAddr: sdk.ConsAddress(pubkey.Address()),
		valAddr:  sdk.ValAddress(simtestutil.CreateRandomAccounts(1)[0]),
		tmPk:     tmPk,
		privKey:  privkey,
	}
}

func (t testValidator) toValidator(power int64) abcitypes.Validator {
	return abcitypes.Validator{
		Address: t.consAddr.Bytes(),
		Power:   power,
	}
}

type ABCITestSuite struct {
	suite.Suite

	vals [3]testValidator
	ctx  sdk.Context

	mockBatch          batchingtypes.Batch
	mockBatchingKeeper *testutil.MockBatchingKeeper
	mockPubKeyKeeper   *testutil.MockPubKeyKeeper
	mockStakingKeeper  *testutil.MockStakingKeeper
	mockTxVerifier     *testutil.MockTxVerifier
}

func (s *ABCITestSuite) SetupSuite() {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	cfg.Seal()
}

func (s *ABCITestSuite) SetupTest(mockBatchNumber uint64, isNewValidator []bool) {
	// Create validators and set up SEDA signer for each of them.
	vals := [3]testValidator{
		newTestValidator(),
		newTestValidator(),
		newTestValidator(),
	}

	tmpDir := s.T().TempDir()
	for i, val := range vals {
		dirPath := filepath.Join(tmpDir, fmt.Sprintf("%d", i))
		err := os.MkdirAll(dirPath, 0755)
		s.Require().NoError(err)

		vals[i].sedaPubKeys, err = utils.GenerateSEDAKeys(val.valAddr, dirPath, "", false)
		s.Require().NoError(err)

		secp256k1PubKey := vals[i].sedaPubKeys[utils.SEDAKeyIndexSecp256k1].PubKey
		vals[i].ethAddr, err = utils.PubKeyToEthAddress(secp256k1PubKey)
		s.Require().NoError(err)

		vals[i].signer, err = utils.LoadSEDASigner(filepath.Join(dirPath, utils.SEDAKeyFileName), true)
		s.Require().NoError(err)

		if isNewValidator != nil && isNewValidator[i] {
			vals[i].isNew = true
		}
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("Message for ECDSA signing"))
	batchID := hasher.Sum(nil)
	mockBatch := batchingtypes.Batch{
		BatchNumber: mockBatchNumber,
		BatchId:     batchID,
		BlockHeight: 100,
	}

	s.mockBatch = mockBatch
	s.vals = vals
	s.ctx = sdk.Context{}.
		WithChainID(chainID).
		WithBlockHeight(mockBatch.BlockHeight).
		WithConsensusParams(cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: 100,
			},
		}).
		WithBlockHeader(cmtproto.Header{
			ChainID: chainID,
			Height:  mockBatch.BlockHeight,
		}).
		WithHeaderInfo(header.Info{
			ChainID: chainID,
			Height:  mockBatch.BlockHeight,
		})

	// Mock configurations (Batch 100 is created at height H)
	ctrl := gomock.NewController(s.T())
	mockBatchingKeeper := testutil.NewMockBatchingKeeper(ctrl)
	mockPubKeyKeeper := testutil.NewMockPubKeyKeeper(ctrl)
	mockStakingKeeper := testutil.NewMockStakingKeeper(ctrl)

	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), mockBatch.BlockHeight).Return(mockBatch, nil).AnyTimes()
	mockBatchingKeeper.EXPECT().GetBatchForHeight(gomock.Any(), mockBatch.BlockHeight+1).Return(batchingtypes.Batch{}, collections.ErrNotFound).AnyTimes()
	for i, val := range s.vals {
		if isNewValidator != nil && isNewValidator[i] {
			mockBatchingKeeper.EXPECT().GetValidatorTreeEntry(gomock.Any(), mockBatch.BatchNumber-1, val.valAddr).
				Return(batchingtypes.ValidatorTreeEntry{}, collections.ErrNotFound).
				AnyTimes()
		} else {
			mockBatchingKeeper.EXPECT().GetValidatorTreeEntry(gomock.Any(), mockBatch.BatchNumber-1, val.valAddr).
				Return(batchingtypes.ValidatorTreeEntry{EthAddress: val.ethAddr}, nil).
				AnyTimes()
		}
		mockPubKeyKeeper.EXPECT().GetValidatorKeys(gomock.Any(), val.valAddr.String()).
			Return(pubkeytypes.ValidatorPubKeys{}, nil).
			AnyTimes()
		mockPubKeyKeeper.EXPECT().GetValidatorKeyAtIndex(gomock.Any(), val.valAddr.Bytes(), utils.SEDAKeyIndexSecp256k1).Return(val.sedaPubKeys[0].PubKey, nil).AnyTimes()

		mockStakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), val.consAddr).
			Return(stakingtypes.Validator{OperatorAddress: val.valAddr.String()}, nil).
			AnyTimes()
	}
	s.mockBatchingKeeper = mockBatchingKeeper
	s.mockPubKeyKeeper = mockPubKeyKeeper
	s.mockStakingKeeper = mockStakingKeeper
	s.mockTxVerifier = testutil.NewMockTxVerifier(ctrl)

	s.mockTxVerifier.EXPECT().ProcessProposalVerifyTx(testutil.ValidTx).Return(testutil.NewMockTx(100), nil).AnyTimes()
	s.mockTxVerifier.EXPECT().ProcessProposalVerifyTx(testutil.InvalidTx).Return(nil, fmt.Errorf("invalid TX")).AnyTimes()
	s.mockTxVerifier.EXPECT().ProcessProposalVerifyTx(testutil.LargeTx).Return(testutil.NewMockTx(10000), nil).AnyTimes()

	// Construct handler for each validator.
	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))
	for i, val := range s.vals {
		defaultProposalHandler := baseapp.NewDefaultProposalHandler(mempool.NewSenderNonceMempool(), s.mockTxVerifier)
		s.vals[i].handlers = abci.NewHandlers(
			defaultProposalHandler.PrepareProposalHandler(),
			defaultProposalHandler.ProcessProposalHandler(),
			mockBatchingKeeper,
			mockPubKeyKeeper,
			mockStakingKeeper,
			authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
			val.signer,
			logger,
		)
	}
}

func (s *ABCITestSuite) incrementBlockHeight() {
	s.ctx = sdk.Context{}.
		WithChainID(s.ctx.ChainID()).
		WithBlockHeight(s.ctx.BlockHeight() + 1).
		WithConsensusParams(cmtproto.ConsensusParams{
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: s.ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight,
			},
			Block: &cmtproto.BlockParams{
				MaxGas: 20000,
			},
		}).
		WithBlockHeader(cmtproto.Header{
			ChainID: s.ctx.ChainID(),
			Height:  s.ctx.BlockHeight() + 1,
		}).
		WithHeaderInfo(header.Info{
			ChainID: s.ctx.ChainID(),
			Height:  s.ctx.BlockHeight() + 1,
		})
}

func (s *ABCITestSuite) validatorVotes(val *testValidator) {
	evRes, err := val.handlers.ExtendVoteHandler()(
		s.ctx, &abcitypes.RequestExtendVote{
			Height: s.ctx.BlockHeight(),
		})
	if val.isNew {
		// New validator does not sign the batch.
		s.Require().Error(err)
		s.Require().Nil(evRes)
	} else {
		s.Require().NoError(err)
		val.voteExt = evRes.VoteExtension
	}
}

func (s *ABCITestSuite) validatorsVote() {
	for i := range s.vals {
		s.validatorVotes(&s.vals[i])
	}
}

func (s *ABCITestSuite) injectCommitInfo(prepareRes *abcitypes.ResponsePrepareProposal, llc abcitypes.ExtendedCommitInfo) {
	injection, err := json.Marshal(llc)
	s.Require().NoError(err)
	prepareRes.Txs = append(prepareRes.Txs, injection)
}

func (s *ABCITestSuite) validatorsProcessProposal(txs [][]byte, expErr string, shouldRejectProposal bool) {
	for _, val := range s.vals {
		processRes, err := val.handlers.ProcessProposalHandler()(
			s.ctx, &abcitypes.RequestProcessProposal{
				ProposedLastCommit: abcitypes.CommitInfo{
					Round: 1,
					Votes: nil,
				},
				Txs:    txs,
				Height: s.ctx.BlockHeight(),
			})
		if expErr != "" {
			s.Require().Error(err)
			s.Require().Contains(err.Error(), expErr)
		}

		if shouldRejectProposal {
			s.Require().Equal(abcitypes.ResponseProcessProposal_REJECT, processRes.Status, "Proposal was accepted when it should be rejected")
			return
		} else {
			s.Require().Equal(abcitypes.ResponseProcessProposal_ACCEPT, processRes.Status, "Proposal was rejected when it should be accepted")
		}
	}
}

func (s *ABCITestSuite) mockExtendedCommitInfo() (abcitypes.ExtendedCommitInfo, comet.BlockInfo) {
	var llc abcitypes.ExtendedCommitInfo
	for i, val := range s.vals {
		s.mockStakingKeeper.EXPECT().GetPubKeyByConsAddr(gomock.Any(), val.consAddr.Bytes()).Return(val.tmPk, nil).AnyTimes()
		s.mockStakingKeeper.EXPECT().GetPubKeyByConsAddr(gomock.Any(), val.consAddr.Bytes()).Return(val.tmPk, nil).AnyTimes()
		s.mockStakingKeeper.EXPECT().GetPubKeyByConsAddr(gomock.Any(), val.consAddr.Bytes()).Return(val.tmPk, nil).AnyTimes()

		cve := cmtproto.CanonicalVoteExtension{
			Extension: val.voteExt,
			Height:    101,
			Round:     int64(0),
			ChainId:   chainID,
		}

		bz, err := marshalDelimitedFn(&cve)
		s.Require().NoError(err)

		extSig, err := val.privKey.Sign(bz)
		s.Require().NoError(err)

		blockIdFlag := cmtproto.BlockIDFlagCommit
		if !val.isNew && len(val.voteExt) == 0 {
			blockIdFlag = cmtproto.BlockIDFlagAbsent
		}

		llc.Votes = append(llc.Votes, abcitypes.ExtendedVoteInfo{
			// Each validator has a slightly different power to avoid ties.
			Validator:          val.toValidator(333 + int64(i)),
			VoteExtension:      val.voteExt,
			ExtensionSignature: extSig,
			BlockIdFlag:        blockIdFlag,
		})
	}

	// Sort and convert to last commit.
	return extendedCommitToLastCommit(llc)
}

func marshalDelimitedFn(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func extendedCommitToLastCommit(ec abcitypes.ExtendedCommitInfo) (abcitypes.ExtendedCommitInfo, comet.BlockInfo) {
	sort.Sort(extendedVoteInfos(ec.Votes))

	// Convert the extended commit info to last commit info.
	lastCommit := abcitypes.CommitInfo{
		Round: ec.Round,
		Votes: make([]abcitypes.VoteInfo, len(ec.Votes)),
	}
	for i, vote := range ec.Votes {
		lastCommit.Votes[i] = abcitypes.VoteInfo{
			Validator: abcitypes.Validator{
				Address: vote.Validator.Address,
				Power:   vote.Validator.Power,
			},
			BlockIdFlag: vote.BlockIdFlag,
		}
	}
	return ec, baseapp.NewBlockInfo(
		nil,
		nil,
		nil,
		lastCommit,
	)
}

type extendedVoteInfos []abcitypes.ExtendedVoteInfo

func (v extendedVoteInfos) Len() int {
	return len(v)
}

func (v extendedVoteInfos) Less(i, j int) bool {
	if v[i].Validator.Power == v[j].Validator.Power {
		return bytes.Compare(v[i].Validator.Address, v[j].Validator.Address) == -1
	}
	return v[i].Validator.Power > v[j].Validator.Power
}

func (v extendedVoteInfos) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
