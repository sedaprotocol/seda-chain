package app

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
)

var _ QueryServer = Querier{}

type Querier struct {
	signer       utils.SEDASigner
	pubKeyKeeper PubKeyKeeper
}

type PubKeyKeeper interface {
	GetValidatorKeyAtIndex(ctx context.Context, validatorAddr sdk.ValAddress, index sedatypes.SEDAKeyIndex) ([]byte, error)
	IsProvingSchemeActivated(ctx context.Context, index sedatypes.SEDAKeyIndex) (bool, error)
}

func NewQuerier(signer utils.SEDASigner, pubKeyKeeper PubKeyKeeper) Querier {
	return Querier{
		signer:       signer,
		pubKeyKeeper: pubKeyKeeper,
	}
}

func (q Querier) SEDASignerStatus(ctx context.Context, _ *QuerySEDASignerStatusRequest) (*QuerySEDASignerStatusResponse, error) {
	pubKeys := q.signer.GetPublicKeys()
	if !q.signer.IsLoaded() {
		return nil, fmt.Errorf("signer is not loaded")
	}

	signerKeys := make([]*SignerKey, len(pubKeys))
	for i, pk := range pubKeys {
		index := sedatypes.SEDAKeyIndex(pk.Index)
		registeredPubKey, err := q.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, q.signer.GetValAddress(), index)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return nil, err
		}

		isActive, err := q.pubKeyKeeper.IsProvingSchemeActivated(ctx, sedatypes.SEDAKeyIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				isActive = false
			} else {
				return nil, err
			}
		}

		signerKeys[i] = &SignerKey{
			Index:                 pk.Index,
			IndexName:             index.String(),
			IsProvingSchemeActive: isActive,
			PublicKey:             hex.EncodeToString(pk.PubKey),
			IsSynced:              bytes.Equal(registeredPubKey, pk.PubKey),
		}
	}

	return &QuerySEDASignerStatusResponse{
		ValidatorAddress: q.signer.GetValAddress().String(),
		SignerKeys:       signerKeys,
	}, nil
}

// GetSEDASignerStatus returns the command for querying the status of
// the node's SEDA signer.
func GetSEDASignerStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seda-signer-status",
		Short: "Query status of the node's SEDA signer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)

			res, err := queryClient.SEDASignerStatus(cmd.Context(), &QuerySEDASignerStatusRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
