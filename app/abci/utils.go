package abci

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/comet"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsVoteExtensionsEnabled determines if vote extensions are enabled for the current block. If
// vote extensions are enabled at height h, then a proposer will receive vote extensions
// in height h+1. This is primarily utilized by any module that needs to make state changes
// based on whether they were included in a proposal.
func IsVoteExtensionsEnabled(ctx sdk.Context) bool {
	cp := ctx.ConsensusParams()
	if cp.Abci == nil || cp.Abci.VoteExtensionsEnableHeight == 0 {
		return false
	}

	// Per the cosmos sdk, the first block should not utilize the latest finalize block state. This means
	// vote extensions should NOT be making state changes.
	//
	// Ref: https://github.com/cosmos/cosmos-sdk/blob/2100a73dcea634ce914977dbddb4991a020ee345/baseapp/baseapp.go#L488-L495
	if ctx.BlockHeight() <= 1 {
		return false
	}

	return cp.Abci.VoteExtensionsEnableHeight < ctx.BlockHeight()
}

// ValidatorStore defines the interface contract require for verifying vote
// extension signatures. Typically, this will be implemented by the x/staking
// module, which has knowledge of the CometBFT public key.
type ValidatorStore interface {
	GetPubKeyByConsAddr(context.Context, sdk.ConsAddress) (cmtprotocrypto.PublicKey, error)
}

// ValidateVoteExtensions defines a helper function for verifying vote extension
// signatures that may be passed or manually injected into a block proposal from
// a proposer in PrepareProposal. It returns an error if any signature is invalid
// or if unexpected vote extensions and/or signatures are found or less than 2/3
// power is received.
func ValidateVoteExtensions(
	ctx sdk.Context,
	valStore ValidatorStore,
	extCommit cometabci.ExtendedCommitInfo,
) error {
	currentHeight := ctx.HeaderInfo().Height
	chainID := ctx.HeaderInfo().ChainID
	commitInfo := ctx.CometInfo().GetLastCommit()

	// Check that both extCommit + commit are ordered in accordance with vp/address.
	if err := ValidateExtendedCommitAgainstLastCommit(extCommit, commitInfo); err != nil {
		return err
	}

	// Start checking vote extensions only **after** the vote extensions enable
	// height, because when `currentHeight == VoteExtensionsEnableHeight`
	// PrepareProposal doesn't get any vote extensions in its request.
	extensionsEnabled := IsVoteExtensionsEnabled(ctx)
	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	var (
		// Total voting power of all vote extensions.
		totalVP int64
		// Total voting power of all validators that submitted valid vote extensions.
		sumVP int64
	)

	for _, vote := range extCommit.Votes {
		totalVP += vote.Validator.Power

		if extensionsEnabled {
			if vote.BlockIdFlag == cmtproto.BlockIDFlagCommit && len(vote.ExtensionSignature) == 0 {
				return fmt.Errorf("vote extension signature is missing; validator addr %s",
					vote.Validator.String(),
				)
			}
			if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit && len(vote.VoteExtension) != 0 {
				return fmt.Errorf("non-commit vote extension present; validator addr %s",
					vote.Validator.String(),
				)
			}
			if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit && len(vote.ExtensionSignature) != 0 {
				return fmt.Errorf("non-commit vote extension signature present; validator addr %s",
					vote.Validator.String(),
				)
			}
		} else { // vote extensions disabled
			if len(vote.VoteExtension) != 0 {
				return fmt.Errorf("vote extension present but extensions disabled; validator addr %s",
					vote.Validator.String(),
				)
			}
			if len(vote.ExtensionSignature) != 0 {
				return fmt.Errorf("vote extension signature present but extensions disabled; validator addr %s",
					vote.Validator.String(),
				)
			}

			continue
		}

		// Only check + include power if the vote is a commit vote. There must be super-majority, otherwise the
		// previous block (the block vote is for) could not have been committed.
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		valConsAddr := sdk.ConsAddress(vote.Validator.Address)

		pubKeyProto, err := valStore.GetPubKeyByConsAddr(ctx, valConsAddr)
		if err != nil {
			continue
		}

		cmtPubKey, err := cryptoenc.PubKeyFromProto(pubKeyProto)
		if err != nil {
			return fmt.Errorf("failed to convert validator %X public key: %w", valConsAddr, err)
		}

		cve := cmtproto.CanonicalVoteExtension{
			Extension: vote.VoteExtension,
			Height:    currentHeight - 1, // the vote extension was signed in the previous height
			Round:     int64(extCommit.Round),
			ChainId:   chainID,
		}

		extSignBytes, err := marshalDelimitedFn(&cve)
		if err != nil {
			return fmt.Errorf("failed to encode CanonicalVoteExtension: %w", err)
		}

		if !cmtPubKey.VerifySignature(extSignBytes, vote.ExtensionSignature) {
			return fmt.Errorf("failed to verify validator %X vote extension signature", valConsAddr)
		}

		sumVP += vote.Validator.Power
	}

	// This check is probably unnecessary, but better safe than sorry.
	if totalVP <= 0 {
		return fmt.Errorf("total voting power must be positive, got: %d", totalVP)
	}

	// If the sum of the voting power has not reached (2/3 + 1) we need to error.
	if requiredVP := ((totalVP * 2) / 3) + 1; sumVP < requiredVP {
		return fmt.Errorf(
			"insufficient cumulative voting power received to verify vote extensions; got: %d, expected: >=%d",
			sumVP, requiredVP,
		)
	}

	return nil
}

// ValidateExtendedCommitAgainstLastCommit validates an ExtendedCommitInfo against a LastCommit. Specifically,
// it checks that the ExtendedCommit + LastCommit (for the same height), are consistent with each other + that
// they are ordered correctly (by voting power) in accordance with
// [comet](https://github.com/cometbft/cometbft/blob/4ce0277b35f31985bbf2c25d3806a184a4510010/types/validator_set.go#L784).
// Modified to not check for block ID flag consistency for absent votes, as those might have been pruned in prepare proposal.
func ValidateExtendedCommitAgainstLastCommit(ec cometabci.ExtendedCommitInfo, lc comet.CommitInfo) error {
	// check that the rounds are the same
	if ec.Round != lc.Round() {
		return fmt.Errorf("extended commit round %d does not match last commit round %d", ec.Round, lc.Round())
	}

	// check that the # of votes are the same
	if len(ec.Votes) != lc.Votes().Len() {
		return fmt.Errorf("extended commit votes length %d does not match last commit votes length %d", len(ec.Votes), lc.Votes().Len())
	}

	// check sort order of extended commit votes
	if !slices.IsSortedFunc(ec.Votes, func(vote1, vote2 cometabci.ExtendedVoteInfo) int {
		if vote1.Validator.Power == vote2.Validator.Power {
			return bytes.Compare(vote1.Validator.Address, vote2.Validator.Address) // addresses sorted in ascending order (used to break vp conflicts)
		}
		return -int(vote1.Validator.Power - vote2.Validator.Power) // vp sorted in descending order
	}) {
		return fmt.Errorf("extended commit votes are not sorted by voting power")
	}

	addressCache := make(map[string]struct{}, len(ec.Votes))
	// check that consistency between LastCommit and ExtendedCommit
	for i, vote := range ec.Votes {
		// cache addresses to check for duplicates
		if _, ok := addressCache[string(vote.Validator.Address)]; ok {
			return fmt.Errorf("extended commit vote address %X is duplicated", vote.Validator.Address)
		}
		addressCache[string(vote.Validator.Address)] = struct{}{}

		lcVote := lc.Votes().Get(i)
		if !bytes.Equal(vote.Validator.Address, lcVote.Validator().Address()) {
			return fmt.Errorf("extended commit vote address %X does not match last commit vote address %X", vote.Validator.Address, lcVote.Validator().Address())
		}
		if vote.Validator.Power != lcVote.Validator().Power() {
			return fmt.Errorf("extended commit vote power %d does not match last commit vote power %d", vote.Validator.Power, lcVote.Validator().Power())
		}

		// only check non-absent votes (these could have been modified via pruning in prepare proposal)
		if !(vote.BlockIdFlag == cmtproto.BlockIDFlagAbsent && len(vote.VoteExtension) == 0 && len(vote.ExtensionSignature) == 0) {
			if int32(vote.BlockIdFlag) != int32(lcVote.GetBlockIDFlag()) {
				return fmt.Errorf("mismatched block ID flag between extended commit vote %d and last proposed commit %d", int32(vote.BlockIdFlag), int32(lcVote.GetBlockIDFlag()))
			}
		}
	}

	return nil
}
