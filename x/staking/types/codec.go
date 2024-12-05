package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInterfaces registers the x/staking interfaces types with
// the interface registry.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateSEDAValidator{},
		&types.MsgEditValidator{},
		&types.MsgDelegate{},
		&types.MsgUndelegate{},
		&types.MsgBeginRedelegate{},
		&types.MsgCancelUnbondingDelegation{},
		&types.MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&types.StakeAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
