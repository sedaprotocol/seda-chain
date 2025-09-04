package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

type Keeper struct {
	bankKeeper      types.BankKeeper
	accountKeeper   types.AccountKeeper
	stakingKeeper   types.StakingKeeper
	dataProxyKeeper types.DataProxyKeeper

	// authority is the address capable of executing MsgUpdateParams. Typically, this should be the gov module address.
	// Initially this will be the SEDA security group address.
	authority string

	Schema             collections.Schema
	params             collections.Item[types.Params]
	fastClientID       collections.Sequence
	fastClient         collections.Map[[]byte, types.FastClient]
	fastUser           collections.Map[collections.Pair[uint64, string], types.FastUser]
	fastClientTransfer collections.Map[uint64, []byte]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	bk types.BankKeeper,
	ak types.AccountKeeper,
	sk types.StakingKeeper,
	dpk types.DataProxyKeeper,
	authority string,
) *Keeper {
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		authority:          authority,
		bankKeeper:         bk,
		accountKeeper:      ak,
		stakingKeeper:      sk,
		dataProxyKeeper:    dpk,
		params:             collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		fastClientID:       collections.NewSequence(sb, types.FastClientIDKey, "fast_client_id"),
		fastClient:         collections.NewMap(sb, types.FastClientKey, "fast_client", collections.BytesKey, codec.CollValue[types.FastClient](cdc)),
		fastUser:           collections.NewMap(sb, types.FastUserKey, "fast_user", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), codec.CollValue[types.FastUser](cdc)),
		fastClientTransfer: collections.NewMap(sb, types.FastClientTransferKey, "fast_client_transfer", collections.Uint64Key, collections.BytesValue),
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

func (k Keeper) setCurrentFastClientID(ctx sdk.Context, fastClientID uint64) error {
	return k.fastClientID.Set(ctx, fastClientID)
}

func (k Keeper) GetCurrentFastClientID(ctx sdk.Context) (uint64, error) {
	return k.fastClientID.Peek(ctx)
}

func (k Keeper) HasFastClient(ctx sdk.Context, pubKey []byte) (bool, error) {
	return k.fastClient.Has(ctx, pubKey)
}

// SetFastClient sets the FastClient for a given public key, meant to be used
// for key rotation or restoring from an export.
func (k Keeper) SetFastClient(ctx sdk.Context, pubKey []byte, fastClient types.FastClient) error {
	return k.fastClient.Set(ctx, pubKey, fastClient)
}

func (k Keeper) DeleteFastClient(ctx sdk.Context, pubKey []byte) error {
	return k.fastClient.Remove(ctx, pubKey)
}

// CreateFastClient creates a new FastClient for a given public key, meant to
// be used for creating a new FastClient.
func (k Keeper) CreateFastClient(ctx sdk.Context, pubKey []byte, fastClientInput types.FastClientInput) (types.FastClient, error) {
	fastClientID, err := k.fastClientID.Next(ctx)
	if err != nil {
		return types.FastClient{}, err
	}

	fastClient := types.FastClient{
		Id:           fastClientID,
		OwnerAddress: fastClientInput.OwnerAddress,
		AdminAddress: fastClientInput.AdminAddress,
		Address:      fastClientInput.Address,
		PublicKey:    fastClientInput.PublicKey,
		Memo:         fastClientInput.Memo,
		Balance:      fastClientInput.Balance,
		UsedCredits:  fastClientInput.UsedCredits,
	}

	if err := k.fastClient.Set(ctx, pubKey, fastClient); err != nil {
		return types.FastClient{}, err
	}

	return fastClient, nil
}

func (k Keeper) GetFastClient(ctx sdk.Context, pubKey []byte) (result types.FastClient, err error) {
	fastClient, err := k.fastClient.Get(ctx, pubKey)
	if err != nil {
		return types.FastClient{}, err
	}

	return fastClient, nil
}

func (k Keeper) HasFastUser(ctx sdk.Context, fastClientID uint64, userID string) (bool, error) {
	return k.fastUser.Has(ctx, collections.Join(fastClientID, userID))
}

func (k Keeper) SetFastUser(ctx sdk.Context, fastClientID uint64, userID string, fastUser types.FastUser) error {
	return k.fastUser.Set(ctx, collections.Join(fastClientID, userID), fastUser)
}

func (k Keeper) DeleteFastUser(ctx sdk.Context, fastClientID uint64, userID string) error {
	return k.fastUser.Remove(ctx, collections.Join(fastClientID, userID))
}

func (k Keeper) GetFastUser(ctx sdk.Context, fastClientID uint64, userID string) (result types.FastUser, err error) {
	fastUser, err := k.fastUser.Get(ctx, collections.Join(fastClientID, userID))
	if err != nil {
		return types.FastUser{}, err
	}

	return fastUser, nil
}

func (k Keeper) GetFastUserByPubKey(ctx sdk.Context, fastClientPubKey []byte, userID string) (result types.FastUser, err error) {
	fastClient, err := k.GetFastClient(ctx, fastClientPubKey)
	if err != nil {
		return types.FastUser{}, err
	}

	return k.GetFastUser(ctx, fastClient.Id, userID)
}

func (k Keeper) GetFastTransfer(ctx sdk.Context, fastClientID uint64) ([]byte, error) {
	return k.fastClientTransfer.Get(ctx, fastClientID)
}

func (k Keeper) SetFastTransfer(ctx sdk.Context, fastClientID uint64, address []byte) error {
	return k.fastClientTransfer.Set(ctx, fastClientID, address)
}

func (k Keeper) DeleteFastTransfer(ctx sdk.Context, fastClientID uint64) error {
	return k.fastClientTransfer.Remove(ctx, fastClientID)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
