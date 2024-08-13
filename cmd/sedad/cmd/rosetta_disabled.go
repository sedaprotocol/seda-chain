//go:build !rosetta

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/sedaprotocol/seda-chain/app"
)

func addRosettaCmd(_ *cobra.Command, _ app.EncodingConfig) {}
