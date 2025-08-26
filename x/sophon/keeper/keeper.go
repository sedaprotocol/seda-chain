package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

type Keeper struct {
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper

	// authority is the address capable of executing MsgUpdateParams. Typically, this should be the gov module address.
	// Initially this will be the SEDA security group address.
	authority string

	Schema         collections.Schema
	params         collections.Item[types.Params]
	sophonID       collections.Sequence
	sophonInfo     collections.Map[[]byte, types.SophonInfo]
	sophonUser     collections.Map[collections.Pair[uint64, []byte], types.SophonUser]
	sophonTransfer collections.KeySet[collections.Pair[uint64, []byte]]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, bk types.BankKeeper, ak types.AccountKeeper, authority string) *Keeper {
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:      authority,
		bankKeeper:     bk,
		accountKeeper:  ak,
		params:         collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		sophonID:       collections.NewSequence(sb, types.SophonIDKey, "sophon_id"),
		sophonInfo:     collections.NewMap(sb, types.SophonInfoKey, "sophon_info", collections.BytesKey, codec.CollValue[types.SophonInfo](cdc)),
		sophonUser:     collections.NewMap(sb, types.SophonUserKey, "sophon_user", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), codec.CollValue[types.SophonUser](cdc)),
		sophonTransfer: collections.NewKeySet(sb, types.SophonTransferKey, "sophon_transfer", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return &k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) setCurrentSophonID(ctx sdk.Context, sophonID uint64) error {
	return k.sophonID.Set(ctx, sophonID)
}

func (k Keeper) GetCurrentSophonID(ctx sdk.Context) (uint64, error) {
	return k.sophonID.Peek(ctx)
}

func (k Keeper) HasSophonInfo(ctx sdk.Context, pubKey []byte) (bool, error) {
	return k.sophonInfo.Has(ctx, pubKey)
}

// SetSophonInfo sets the SophonInfo for a given public key, meant to be used
// for key rotation or restoring from an export.
func (k Keeper) SetSophonInfo(ctx sdk.Context, pubKey []byte, sophonInfo types.SophonInfo) error {
	return k.sophonInfo.Set(ctx, pubKey, sophonInfo)
}

// CreateSophonInfo creates a new SophonInfo for a given public key, meant to
// be used for creating a new Sophon.
func (k Keeper) CreateSophonInfo(ctx sdk.Context, pubKey []byte, sophonInput types.SophonInputs) (types.SophonInfo, error) {
	sophonID, err := k.sophonID.Next(ctx)
	if err != nil {
		return types.SophonInfo{}, err
	}

	sophonInfo := types.SophonInfo{
		Id:           sophonID,
		OwnerAddress: sophonInput.OwnerAddress,
		AdminAddress: sophonInput.AdminAddress,
		Address:      sophonInput.Address,
		PublicKey:    sophonInput.PublicKey,
		Memo:         sophonInput.Memo,
		Balance:      sophonInput.Balance,
		UsedCredits:  sophonInput.UsedCredits,
	}

	if err := k.sophonInfo.Set(ctx, pubKey, sophonInfo); err != nil {
		return types.SophonInfo{}, err
	}

	return sophonInfo, nil
}

func (k Keeper) GetSophonInfo(ctx sdk.Context, pubKey []byte) (result types.SophonInfo, err error) {
	sophonInfo, err := k.sophonInfo.Get(ctx, pubKey)
	if err != nil {
		return types.SophonInfo{}, err
	}

	return sophonInfo, nil
}

func (k Keeper) HasSophonUser(ctx sdk.Context, sophonID uint64, userID []byte) (bool, error) {
	return k.sophonUser.Has(ctx, collections.Join(sophonID, userID))
}

func (k Keeper) SetSophonUser(ctx sdk.Context, sophonID uint64, userID []byte, sophonUser types.SophonUser) error {
	return k.sophonUser.Set(ctx, collections.Join(sophonID, userID), sophonUser)
}

func (k Keeper) GetSophonUser(ctx sdk.Context, sophonID uint64, userID []byte) (result types.SophonUser, err error) {
	sophonUser, err := k.sophonUser.Get(ctx, collections.Join(sophonID, userID))
	if err != nil {
		return types.SophonUser{}, err
	}

	return sophonUser, nil
}

func (k Keeper) GetSophonUserByPubKey(ctx sdk.Context, sophonPubKey []byte, userID []byte) (result types.SophonUser, err error) {
	sophonInfo, err := k.GetSophonInfo(ctx, sophonPubKey)
	if err != nil {
		return types.SophonUser{}, err
	}

	return k.GetSophonUser(ctx, sophonInfo.Id, userID)
}

func (k Keeper) HasSophonTransfer(ctx sdk.Context, sophonID uint64, address []byte) (bool, error) {
	return k.sophonTransfer.Has(ctx, collections.Join(sophonID, address))
}

func (k Keeper) SetSophonTransfer(ctx sdk.Context, sophonID uint64, address []byte) error {
	return k.sophonTransfer.Set(ctx, collections.Join(sophonID, address))
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
