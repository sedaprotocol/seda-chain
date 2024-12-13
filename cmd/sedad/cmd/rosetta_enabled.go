//go:build rosetta

package cmd

import (
	"github.com/spf13/cobra"

	rosettaCmd "github.com/cosmos/rosetta/cmd"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func addRosettaCmd(cmd *cobra.Command, interfaceRegistry codectypes.InterfaceRegistry, cdc codec.Codec) {
	cmd.AddCommand(rosettaCmd.RosettaCommand(interfaceRegistry, cdc))
}
