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

type Info group.GroupInfo

func (p Info) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

type Member group.GroupMember

func (p Member) MarshalJSON() ([]byte, error) {
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

		return types.NewMessage("group-info", Info(groupInfo), ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, []byte{groupkeeper.GroupMemberTablePrefix, 0}); found {
		if change.Delete {
			groupID := storetypes.BigEndianToUint64(keyBytes[:8])
			err := sdk.VerifyAddressFormat(keyBytes[8:])
			if err != nil {
				return nil, err
			}
			data := struct {
				GroupID       string `json:"groupID"`
				MemberAddress string `json:"memberAddress"`
			}{
				GroupID:       fmt.Sprintf("%d", groupID),
				MemberAddress: sdk.AccAddress(keyBytes[8:]).String(),
			}

			return types.NewMessage("group-member-delete", data, ctx), nil
		}

		var member group.GroupMember
		err := cdc.Unmarshal(change.Value, &member)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("group-member", Member(member), ctx), nil
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

		var decisionPolicy json.RawMessage
		if policy.DecisionPolicy != nil {
			policyJSON, err := cdc.MarshalJSON(policy.DecisionPolicy)
			if err != nil {
				// Fallback to just type URL if marshaling fails
				decisionPolicy = json.RawMessage(fmt.Sprintf(`{"@type":"%s"}`, policy.DecisionPolicy.GetTypeUrl()))
			} else {
				decisionPolicy = policyJSON
			}
		}

		data := struct {
			Address        string          `json:"address"`
			GroupID        uint64          `json:"group_id"`
			Admin          string          `json:"admin"`
			Metadata       string          `json:"metadata"`
			Version        uint64          `json:"version"`
			DecisionPolicy json.RawMessage `json:"decision_policy,omitempty"`
			CreatedAt      string          `json:"created_at"`
		}{
			Address:        policy.Address,
			GroupID:        policy.GroupId,
			Admin:          policy.Admin,
			Metadata:       policy.Metadata,
			Version:        policy.Version,
			DecisionPolicy: decisionPolicy,
			CreatedAt:      policy.CreatedAt.String(),
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
	}
	logger.Trace("skipping change", "change", change)
	return nil, nil
}
