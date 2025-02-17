package abci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/sha3"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cometabci "github.com/cometbft/cometbft/abci/types"
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
	signer      utils.SEDASigner
	handlers    *Handlers
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

	// Construct handler for each validator.
	buf := &bytes.Buffer{}
	logger := log.NewLogger(buf, log.LevelOption(zerolog.DebugLevel))
	for i, val := range s.vals {
		defaultProposalHandler := baseapp.NewDefaultProposalHandler(mempool.NoOpMempool{}, nil)
		s.vals[i].handlers = NewHandlers(
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

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}

func (s *ABCITestSuite) TestABCIHandlers() {
	testCases := []struct {
		name               string
		mockBatchNumber    uint64
		heightWithoutBatch bool
		isNewValidator     []bool
		emptyVoteExt       []bool
		overwriteVoteExt   [][]byte
		expErr             string
	}{
		{
			name:            "happy path",
			mockBatchNumber: 100,
		},
		{
			name:               "no batch to sign",
			mockBatchNumber:    100,
			heightWithoutBatch: true,
		},
		{
			name:            "happy path with new validator",
			mockBatchNumber: 100,
			isNewValidator:  []bool{false, false, true},
		},
		{
			name:            "one empty vote extension",
			mockBatchNumber: 100,
			emptyVoteExt:    []bool{false, false, true},
		},
		{
			name:            "first batch",
			mockBatchNumber: collections.DefaultSequenceStart,
		},
		{
			name:             "unrecoverable signature is injected in proposal",
			mockBatchNumber:  100,
			overwriteVoteExt: [][]byte{bytes.Repeat([]byte("b"), 65), nil, nil},
			expErr:           "invalid signature recovery id",
		},
		{
			name:             "invalid signature is injected in proposal",
			mockBatchNumber:  100,
			overwriteVoteExt: [][]byte{nil, nil, {189, 92, 197, 8, 100, 52, 95, 183, 251, 111, 24, 99, 59, 203, 64, 250, 13, 35, 168, 193, 106, 244, 191, 48, 10, 108, 68, 197, 222, 59, 230, 110, 21, 108, 12, 217, 108, 92, 115, 214, 255, 70, 107, 170, 228, 54, 53, 157, 41, 140, 40, 132, 157, 197, 248, 219, 113, 227, 148, 194, 197, 46, 153, 49, 0}},
			expErr:           "batch signature is invalid",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(tc.mockBatchNumber, tc.isNewValidator)

			s.incrementBlockHeight()
			if tc.heightWithoutBatch {
				s.incrementBlockHeight()
			}
			for i, val := range s.vals {
				// ExtendVote at H+1
				evRes, err := val.handlers.ExtendVoteHandler()(
					s.ctx, &cometabci.RequestExtendVote{
						Height: s.ctx.BlockHeight(),
					})
				if tc.isNewValidator != nil && tc.isNewValidator[i] {
					// New validator does not sign the batch.
					s.Require().Error(err)
					s.Require().Nil(evRes)
				} else {
					s.Require().NoError(err)

					// Recover and verify public key.
					if !tc.heightWithoutBatch {
						sigPubKey, err := crypto.Ecrecover(s.mockBatch.BatchId, evRes.VoteExtension)
						s.Require().NoError(err)
						s.Require().Equal(val.sedaPubKeys[utils.SEDAKeyIndexSecp256k1].PubKey, sigPubKey)

						s.vals[i].voteExt = evRes.VoteExtension
						if tc.emptyVoteExt != nil && tc.emptyVoteExt[i] {
							s.vals[i].voteExt = nil
						} else if tc.overwriteVoteExt != nil && tc.overwriteVoteExt[i] != nil {
							s.vals[i].voteExt = tc.overwriteVoteExt[i]
						}
					}
				}

				// VerifyVoteExtension at H+1 (all other validators)
				for _, otherVal := range s.vals {
					if otherVal.consAddr.Equals(val.consAddr) {
						continue
					}
					vvRes, err := otherVal.handlers.VerifyVoteExtensionHandler()(
						s.ctx, &cometabci.RequestVerifyVoteExtension{
							Height:           s.ctx.BlockHeight(),
							VoteExtension:    s.vals[i].voteExt,
							ValidatorAddress: val.consAddr,
						})
					if (tc.emptyVoteExt != nil && tc.emptyVoteExt[i]) ||
						(tc.overwriteVoteExt != nil && tc.overwriteVoteExt[i] != nil) {
						s.Require().Error(err)
						s.Require().Contains(err.Error(), tc.expErr)
					} else {
						s.Require().NoError(err)
						s.Require().Equal(cometabci.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
					}
				}
			}

			// PrepareProposal at H+2 (first validator)
			s.incrementBlockHeight()

			llc, info := s.mockExtendedCommitInfo()
			s.ctx = s.ctx.WithCometInfo(info)

			prepareRes, err := s.vals[0].handlers.PrepareProposalHandler()(
				s.ctx, &cometabci.RequestPrepareProposal{
					LocalLastCommit: llc,
					MaxTxBytes:      22020096,
					Height:          s.ctx.BlockHeight(),
				})
			s.Require().NoError(err)

			// Ensure last commit was injected.
			if !tc.heightWithoutBatch {
				var injected abcitypes.ExtendedCommitInfo
				err = json.Unmarshal(prepareRes.Txs[0], &injected)
				s.Require().NoError(err)
				s.Require().Equal(llc, injected)
			} else {
				s.Require().Equal(0, len(prepareRes.Txs))
			}

			// ProcessProposal at H+2 (all validators)
			for _, val := range s.vals {
				processRes, err := val.handlers.ProcessProposalHandler()(
					s.ctx, &cometabci.RequestProcessProposal{
						ProposedLastCommit: cometabci.CommitInfo{
							Round: 1,
							Votes: nil,
						},
						Txs:    prepareRes.Txs,
						Height: s.ctx.BlockHeight(),
					})

				if tc.emptyVoteExt != nil || tc.overwriteVoteExt != nil {
					s.Require().Error(err)
					s.Require().Contains(err.Error(), tc.expErr)
					return
				} else {
					s.Require().NoError(err)
					s.Require().Equal(cometabci.ResponseProcessProposal_ACCEPT, processRes.Status)
				}
			}

			// PreBlocker at H+2 (all validators)
			if !tc.heightWithoutBatch {
				for _, val := range s.vals {
					s.mockBatchingKeeper.EXPECT().SetBatchSigSecp256k1(gomock.Any(), s.mockBatch.BatchNumber, val.valAddr, val.voteExt).Return(nil).Times(len(s.vals))
				}
			}
			for _, val := range s.vals {
				_, err := val.handlers.PreBlocker()(
					s.ctx, &cometabci.RequestFinalizeBlock{
						Txs:    prepareRes.Txs,
						Height: s.ctx.BlockHeight(),
					})
				s.Require().NoError(err)
			}
		})
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

func (s *ABCITestSuite) mockExtendedCommitInfo() (abcitypes.ExtendedCommitInfo, comet.BlockInfo) {
	var llc abcitypes.ExtendedCommitInfo
	for _, val := range s.vals {
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

		llc.Votes = append(llc.Votes, abcitypes.ExtendedVoteInfo{
			Validator:          val.toValidator(333),
			VoteExtension:      val.voteExt,
			ExtensionSignature: extSig,
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
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
