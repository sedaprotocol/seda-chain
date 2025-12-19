package group

import (
	"bytes"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const StoreKey = group.StoreKey

type GroupInfo group.GroupInfo

func (p GroupInfo) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

type GroupMember group.GroupMember

func (p GroupMember) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}



type Vote group.Vote

func (p Vote) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	logger.Debug("extracting group update", "change", change)

	if _, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.GroupTablePrefix, 0}); found {
		var groupInfo group.GroupInfo
		err := cdc.Unmarshal(change.Value, &groupInfo)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("group-info", GroupInfo(groupInfo), ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.GroupMemberTablePrefix, 0}); found {
		if change.Delete {
			groupID := storetypes.BigEndianToUint64(keyBytes[:8])
			var memberAddress sdk.AccAddress
			if err := memberAddress.Unmarshal(keyBytes[8:]); err != nil {
				return nil, err
			}

			data := struct {
				GroupID       string `json:"groupID"`
				MemberAddress string `json:"memberAddress"`
			}{
				GroupID:       fmt.Sprintf("%d", groupID),
				MemberAddress: memberAddress.String(),
			}

			return types.NewMessage("group-member-delete", data, ctx), nil
		}

		var member group.GroupMember
		err := cdc.Unmarshal(change.Value, &member)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("group-member", GroupMember(member), ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.GroupPolicyTablePrefix, 0}); found {
		if change.Delete {
			logger.Trace("skipping group-policy delete", "change", change)
			return nil, nil
		}

		var policy group.GroupPolicyInfo
		err := cdc.Unmarshal(change.Value, &policy)
		if err != nil {
			return nil, err
		}

		// Use simple struct to avoid Any type marshaling issues
		data := struct {
			Address  string `json:"address"`
			GroupID  uint64 `json:"group_id"`
			Admin    string `json:"admin"`
			Metadata string `json:"metadata"`
			Version  uint64 `json:"version"`
		}{
			Address:  policy.Address,
			GroupID:  policy.GroupId,
			Admin:    policy.Admin,
			Metadata: policy.Metadata,
			Version:  policy.Version,
		}

		return types.NewMessage("group-policy", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.ProposalTablePrefix, 0}); found {
		if change.Delete {
			proposalID := storetypes.BigEndianToUint64(keyBytes[:8])

			data := struct {
				ProposalID string `json:"proposalID"`
			}{
				ProposalID: fmt.Sprintf("%d", proposalID),
			}

			return types.NewMessage("group-proposal-delete", data, ctx), nil
		}

		var proposal group.Proposal
		err := cdc.Unmarshal(change.Value, &proposal)
		if err != nil {
			return nil, err
		}

		// Decode messages using codec to get proper JSON
		messages := make([]json.RawMessage, len(proposal.Messages))
		for i, msg := range proposal.Messages {
			msgJSON, err := cdc.MarshalJSON(msg)
			if err != nil {
				// Fallback to just type URL if marshaling fails
				messages[i] = json.RawMessage(fmt.Sprintf(`{"@type":"%s"}`, msg.GetTypeUrl()))
			} else {
				messages[i] = msgJSON
			}
		}

		// Use simple struct with decoded messages
		data := struct {
			ID                 uint64            `json:"id"`
			GroupPolicyAddress string            `json:"group_policy_address"`
			Metadata           string            `json:"metadata"`
			Proposers          []string          `json:"proposers"`
			SubmitTime         string            `json:"submit_time"`
			GroupVersion       uint64            `json:"group_version"`
			GroupPolicyVersion uint64            `json:"group_policy_version"`
			Status             string            `json:"status"`
			VotingPeriodEnd    string            `json:"voting_period_end"`
			ExecutorResult     string            `json:"executor_result"`
			Title              string            `json:"title"`
			Summary            string            `json:"summary"`
			Messages           []json.RawMessage `json:"messages"`
		}{
			ID:                 proposal.Id,
			GroupPolicyAddress: proposal.GroupPolicyAddress,
			Metadata:           proposal.Metadata,
			Proposers:          proposal.Proposers,
			SubmitTime:         proposal.SubmitTime.String(),
			GroupVersion:       proposal.GroupVersion,
			GroupPolicyVersion: proposal.GroupPolicyVersion,
			Status:             proposal.Status.String(),
			VotingPeriodEnd:    proposal.VotingPeriodEnd.String(),
			ExecutorResult:     proposal.ExecutorResult.String(),
			Title:              proposal.Title,
			Summary:            proposal.Summary,
			Messages:           messages,
		}

		return types.NewMessage("group-proposal", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.VoteTablePrefix, 0}); found {
		if change.Delete {
			return nil, nil
		}

		var vote group.Vote
		err := cdc.Unmarshal(change.Value, &vote)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("group-vote", Vote(vote), ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}