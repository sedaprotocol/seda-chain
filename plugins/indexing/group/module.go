package group

import (
	"bytes"
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
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
