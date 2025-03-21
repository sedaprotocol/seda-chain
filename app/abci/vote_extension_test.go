package abci_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/app/abci"
	"github.com/sedaprotocol/seda-chain/app/abci/testutil"
)

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}

func (s *ABCITestSuite) TestABCIHandlersVerifyVoteExtensionHandler() {
	testCases := []struct {
		name               string
		mockBatchNumber    uint64
		isNewValidator     []bool
		heightWithoutBatch bool
		reqVoteExt         *abcitypes.RequestVerifyVoteExtension
		expectedErr        string
		shouldReject       bool
	}{
		{
			name:            "new batch + valid signature",
			mockBatchNumber: 100,
		},
		{
			name:               "no batch to sign",
			heightWithoutBatch: true,
		},
		{
			name:            "new validator",
			mockBatchNumber: 100,
			isNewValidator:  []bool{false, true, false},
		},
		{
			name:            "new batch + short signature",
			mockBatchNumber: 100,
			reqVoteExt:      &abcitypes.RequestVerifyVoteExtension{VoteExtension: []byte("invalid")},
			expectedErr:     "vote extension is too short",
			shouldReject:    true,
		},
		{
			name:            "new batch + long signature",
			mockBatchNumber: 100,
			reqVoteExt:      &abcitypes.RequestVerifyVoteExtension{VoteExtension: make([]byte, abci.MaxVoteExtensionLength+1)},
			expectedErr:     "vote extension exceeds max length",
			shouldReject:    true,
		},
		{
			name:            "new batch + invalid signature",
			mockBatchNumber: 100,
			reqVoteExt:      &abcitypes.RequestVerifyVoteExtension{VoteExtension: make([]byte, abci.MaxVoteExtensionLength)},
			expectedErr:     "recovery failed",
			shouldReject:    true,
		},
		{
			name:               "no batch + signature",
			heightWithoutBatch: true,
			reqVoteExt:         &abcitypes.RequestVerifyVoteExtension{VoteExtension: []byte("garbage")},
			shouldReject:       true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(tc.mockBatchNumber, tc.isNewValidator)

			s.incrementBlockHeight()
			if tc.heightWithoutBatch {
				s.incrementBlockHeight()
			}

			// Validator 0 extends the vote.
			if tc.reqVoteExt == nil {
				s.validatorVotes(&s.vals[0])
			} else {
				s.vals[0].voteExt = tc.reqVoteExt.VoteExtension
			}

			// Validator 1 verifies the vote.
			vvRes, err := s.vals[1].handlers.VerifyVoteExtensionHandler()(
				s.ctx, &abcitypes.RequestVerifyVoteExtension{
					Height:           s.ctx.BlockHeight(),
					VoteExtension:    s.vals[0].voteExt,
					ValidatorAddress: s.vals[0].consAddr,
				})
			if tc.shouldReject {
				s.Require().Equal(abcitypes.ResponseVerifyVoteExtension_REJECT, vvRes.Status)
				if tc.expectedErr != "" {
					s.Require().Error(err)
					s.Require().Contains(err.Error(), tc.expectedErr)
				}
			} else {
				s.Require().NoError(err)
				s.Require().Equal(abcitypes.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
			}
		})
	}
}

type honestTestCase struct {
	name               string
	mockBatchNumber    uint64
	heightWithoutBatch bool
	isNewValidator     []bool
	// If set, the vote extension is not verified simulating a vote received after the block was delivered to the
	// application and bypassing the VerifyVoteExtensionHandler.
	voteExts           []*abcitypes.ResponseExtendVote
	additionalTxBytes  [][]byte
	expectedPrepareErr string
}

func (tc *honestTestCase) IsNewValidator(i int) bool {
	if tc.isNewValidator == nil {
		return false
	}

	return tc.isNewValidator[i]
}

func (tc *honestTestCase) IsUnverifiedVoteExtension(i int) bool {
	if tc.voteExts == nil {
		return false
	}

	return tc.voteExts[i] != nil
}

func (tc *honestTestCase) GetVoteExtOverwrite(i int) *abcitypes.ResponseExtendVote {
	if tc.voteExts == nil {
		return nil
	}

	return tc.voteExts[i]
}

func (s *ABCITestSuite) TestABCIHandlersHonestProposer() {
	testCases := []honestTestCase{
		{
			name:            "first batch",
			mockBatchNumber: collections.DefaultSequenceStart,
		},
		{
			name:            "happy path without transactions",
			mockBatchNumber: 100,
		},
		{
			name:            "happy path with transactions",
			mockBatchNumber: 100,
			additionalTxBytes: [][]byte{
				testutil.ValidTx,
				testutil.ValidTx,
				testutil.ValidTx,
			},
		},
		{
			name:               "no batch to sign",
			mockBatchNumber:    100,
			heightWithoutBatch: true,
		},
		{
			name:            "new validators",
			mockBatchNumber: 100,
			isNewValidator:  []bool{false, false, true},
		},
		{
			name:            "new validators",
			mockBatchNumber: 100,
			isNewValidator:  []bool{true, false, true},
		},
		{
			name:            "one absent vote of less than 1/3",
			mockBatchNumber: 100,
			voteExts:        []*abcitypes.ResponseExtendVote{{VoteExtension: []byte{}}, nil, nil},
		},
		{
			name:               "absent votes of more than 1/3",
			mockBatchNumber:    100,
			voteExts:           []*abcitypes.ResponseExtendVote{nil, nil, {VoteExtension: []byte{}}},
			expectedPrepareErr: "insufficient cumulative voting power received to verify vote extensions;",
		},
		{
			name:            "unverified vote extensions of less than 1/3",
			mockBatchNumber: 100,
			voteExts:        []*abcitypes.ResponseExtendVote{{VoteExtension: []byte("this is not a valid vote extension")}, nil, nil},
		},
		{
			name:               "unverified vote extensions of more than 1/3",
			mockBatchNumber:    100,
			voteExts:           []*abcitypes.ResponseExtendVote{nil, nil, {VoteExtension: []byte("this is not a valid vote extension")}},
			expectedPrepareErr: "insufficient cumulative voting power received to verify vote extensions;",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(tc.mockBatchNumber, tc.isNewValidator)

			s.incrementBlockHeight()
			if tc.heightWithoutBatch {
				s.incrementBlockHeight()
			}

			if !tc.heightWithoutBatch {
				for i, val := range s.vals {
					// ExtendVote at H+1
					evRes := tc.GetVoteExtOverwrite(i)
					if evRes == nil {
						s.validatorVotes(&s.vals[i])
					} else {
						s.vals[i].voteExt = evRes.VoteExtension
					}

					// VerifyVoteExtension at H+1
					for _, otherVal := range s.vals {
						if otherVal.consAddr.Equals(val.consAddr) || tc.IsUnverifiedVoteExtension(i) {
							continue
						}
						vvRes, err := otherVal.handlers.VerifyVoteExtensionHandler()(
							s.ctx, &abcitypes.RequestVerifyVoteExtension{
								Height:           s.ctx.BlockHeight(),
								VoteExtension:    s.vals[i].voteExt,
								ValidatorAddress: val.consAddr,
							})
						s.Require().NoError(err)
						s.Require().Equal(abcitypes.ResponseVerifyVoteExtension_ACCEPT, vvRes.Status)
					}
				}
			}

			// PrepareProposal at H+2 (first validator)
			s.incrementBlockHeight()

			llc, info := s.mockExtendedCommitInfo()
			s.ctx = s.ctx.WithCometInfo(info)

			prepareRes, err := s.vals[0].handlers.PrepareProposalHandler()(
				s.ctx, &abcitypes.RequestPrepareProposal{
					LocalLastCommit: llc,
					MaxTxBytes:      22020096,
					Height:          s.ctx.BlockHeight(),
				})
			if tc.expectedPrepareErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedPrepareErr)
				return
			}
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

			if tc.additionalTxBytes != nil {
				prepareRes.Txs = append(prepareRes.Txs, tc.additionalTxBytes...)
			}

			// ProcessProposal at H+2 (all validators)
			s.validatorsProcessProposal(prepareRes.Txs, "", false)

			// PreBlocker at H+2 (all validators)
			if !tc.heightWithoutBatch {
				for i, val := range s.vals {
					if tc.IsUnverifiedVoteExtension(i) || tc.IsNewValidator(i) {
						continue
					}
					s.mockBatchingKeeper.EXPECT().SetBatchSigSecp256k1(gomock.Any(), s.mockBatch.BatchNumber, val.valAddr, val.voteExt).Return(nil).Times(len(s.vals))
				}
			}
			for _, val := range s.vals {
				_, err := val.handlers.PreBlocker()(
					s.ctx, &abcitypes.RequestFinalizeBlock{
						Txs:    prepareRes.Txs,
						Height: s.ctx.BlockHeight(),
					})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ABCITestSuite) maliciousSetup() {
	s.SetupTest(100, []bool{false, false, false})

	// ExtendVote at H+1
	s.incrementBlockHeight()
	s.validatorsVote()

	// PrepareProposal at H+2
	s.incrementBlockHeight()
}

func (s *ABCITestSuite) TestABCIHandlersMaliciousProposer() {
	// Unfortunately there is no way to prevent pruning valid votes.
	s.Run("proposer prunes 1/3rd of the votes despite being valid", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[2].BlockIdFlag = cmtproto.BlockIDFlagAbsent
		llc.Votes[2].VoteExtension = nil
		llc.Votes[2].ExtensionSignature = nil

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "", false)

		// PreBlocker at H+2 (all validators)
		for _, val := range s.vals {
			if bytes.Equal(val.consAddr.Bytes(), llc.Votes[2].Validator.Address) {
				continue
			}
			s.mockBatchingKeeper.EXPECT().SetBatchSigSecp256k1(gomock.Any(), s.mockBatch.BatchNumber, val.valAddr, val.voteExt).Return(nil).Times(len(s.vals))
		}
		for _, val := range s.vals {
			_, err := val.handlers.PreBlocker()(
				s.ctx, &abcitypes.RequestFinalizeBlock{
					Txs:    prepareRes.Txs,
					Height: s.ctx.BlockHeight(),
				})
			s.Require().NoError(err)
		}
	})

	s.Run("proposer prunes more than 1/3rd of the votes", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[0].BlockIdFlag = cmtproto.BlockIDFlagAbsent
		llc.Votes[0].VoteExtension = nil
		llc.Votes[0].ExtensionSignature = nil

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "insufficient cumulative voting power received to verify vote extensions; got: 667, expected: >=66", true)
	})

	s.Run("proposer marks a vote as absent", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[0].BlockIdFlag = cmtproto.BlockIDFlagAbsent

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "mismatched block ID flag between extended commit vote 1 and last proposed commit", true)
	})

	s.Run("proposer injects an invalid vote extension which should have been pruned", func() {
		s.maliciousSetup()

		// Make the vote extension invalid, the call to mockExtendedCommitInfo will sign it.
		s.vals[0].voteExt = bytes.Repeat([]byte{0x01}, 65)

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "batch signature is invalid", true)
	})

	s.Run("proposer empties a vote extension", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[0].VoteExtension = []byte{}

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, fmt.Sprintf("failed to verify validator %X vote extension signature", llc.Votes[0].Validator.Address), true)
	})

	s.Run("proposer manipulates a vote", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[0].VoteExtension = []byte("invalid")

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, fmt.Sprintf("failed to verify validator %X vote extension signature", llc.Votes[0].Validator.Address), true)
	})

	s.Run("proposer injects votes without batch", func() {
		s.maliciousSetup()

		s.incrementBlockHeight()
		// The default process proposal handler should reject the proposal as it's an invalid tx.
		s.mockTxVerifier.EXPECT().ProcessProposalVerifyTx(gomock.Any()).Return(nil, fmt.Errorf("invalid TX")).AnyTimes()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "", true)
	})

	s.Run("proposer adds an additional vote", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes = append(llc.Votes, abcitypes.ExtendedVoteInfo{
			Validator: abcitypes.Validator{
				Address: mustDecodeBase64("a5XdK0N65IhetG24u0rbKjHMcEo="),
				Power:   400,
			},
			VoteExtension:      mustDecodeBase64("DSyP4ZXSX9T2s0YBAQl2GPcUpZ4pKdeFObjklK5mnSkxHxvQW51whL8ubsckGbWHNZpLP1b102VG6YZfrO/mcwE="),
			ExtensionSignature: mustDecodeBase64("l42SpFcnbqi8EAdWeaN5KSLvF2zfTW2MXjwxoB6WweoIG3a2HHC9PP5kq3Ox1M452HBHZS2j6oDeG+OqwYmxsA=="),
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
		})

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "extended commit votes length 4 does not match last commit votes length 3", true)
	})

	s.Run("proposer removes a vote", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes = llc.Votes[1:]

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "extended commit votes length 2 does not match last commit votes length 3", true)
	})

	s.Run("proposer replaces a vote", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		llc.Votes[0] = abcitypes.ExtendedVoteInfo{
			Validator: abcitypes.Validator{
				Address: mustDecodeBase64("a5XdK0N65IhetG24u0rbKjHMcEo="),
				Power:   400,
			},
			VoteExtension:      mustDecodeBase64("DSyP4ZXSX9T2s0YBAQl2GPcUpZ4pKdeFObjklK5mnSkxHxvQW51whL8ubsckGbWHNZpLP1b102VG6YZfrO/mcwE="),
			ExtensionSignature: mustDecodeBase64("l42SpFcnbqi8EAdWeaN5KSLvF2zfTW2MXjwxoB6WweoIG3a2HHC9PP5kq3Ox1M452HBHZS2j6oDeG+OqwYmxsA=="),
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
		}

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)

		s.validatorsProcessProposal(prepareRes.Txs, "extended commit vote address 6B95DD2B437AE4885EB46DB8BB4ADB2A31CC704A does not match last commit vote address", true)
	})

	s.Run("proposer bloats the proposal with invalid txs", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)
		prepareRes.Txs = append(prepareRes.Txs, testutil.ValidTx, testutil.ValidTx, testutil.InvalidTx)

		s.validatorsProcessProposal(prepareRes.Txs, "", true)
	})

	s.Run("proposer exceeds the block gas limit", func() {
		s.maliciousSetup()

		llc, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		prepareRes := abcitypes.ResponsePrepareProposal{}
		s.injectCommitInfo(&prepareRes, llc)
		prepareRes.Txs = append(prepareRes.Txs, testutil.LargeTx, testutil.LargeTx, testutil.LargeTx)

		s.validatorsProcessProposal(prepareRes.Txs, "", true)
	})

	s.Run("proposer does not inject extended votes", func() {
		s.maliciousSetup()

		_, info := s.mockExtendedCommitInfo()
		s.ctx = s.ctx.WithCometInfo(info)

		prepareRes := abcitypes.ResponsePrepareProposal{}

		s.validatorsProcessProposal(prepareRes.Txs, "no injected extended votes tx", true)
	})
}
