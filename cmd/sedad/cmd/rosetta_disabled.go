//go:build !rosetta

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func addRosettaCmd(cmd *cobra.Command, _ codectypes.InterfaceRegistry, _ codec.Codec) {}
