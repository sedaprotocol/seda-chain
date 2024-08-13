//go:build rosetta

package cmd

import (
	rosettaCmd "github.com/cosmos/rosetta/cmd"
	"github.com/spf13/cobra"

	"github.com/sedaprotocol/seda-chain/app"
)

func addRosettaCmd(cmd *cobra.Command, encodingCfg app.EncodingConfig) {
	cmd.AddCommand(rosettaCmd.RosettaCommand(encodingCfg.InterfaceRegistry, encodingCfg.Marshaler))
}
