package keeper

import (
	"errors"
	"strings"

	"cosmossdk.io/api/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto/types"
)

func (k Keeper) AddKey(ctx sdk.Context, cmd *cobra.Command, name, pass, application string, coinType, account, index uint32, algo keyring.SignatureAlgo) (sdkcrypto.PubKey, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("the provided name is invalid or empty after trimming whitespace")
	}

	if ok, _ := k.KeyName.Has(ctx); ok {
		return nil, errors.New("key already exists")
	}

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return nil, err
	}
	kb := clientCtx.Keyring

	if err := k.KeyName.Set(ctx, name); err != nil {
		return nil, err
	}
	hdPath := hd.CreateHDPath(coinType, account, index).String()

	// read entropy seed straight from cmtcrypto.Rand and convert to mnemonic
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}

	// Read mnemonic from horcrux or tmkms.
	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return nil, err
	}

	r, err := kb.NewAccount(name, mnemonic, pass, hdPath, algo)
	if err != nil {
		return nil, err
	}
	key, err := r.GetPubKey()
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (k Keeper) GetAllKeys(ctx sdk.Context) ([]crypto.PublicKey, error) {
	return nil, nil
}
