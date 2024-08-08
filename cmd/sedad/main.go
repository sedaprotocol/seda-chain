package main

import (
	"fmt"
	"os"

	//"cosmossdk.io/log"
	//svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	_ "github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-wasm-vm/tallyvm"
	//"github.com/sedaprotocol/seda-chain/cmd/sedad/cmd"
)

func main() {
	fmt.Println("starting program")

	tallyWasm, err := os.ReadFile("./testwasm.wasm")
	if err != nil {
		panic(err)
	}

	vmRes := tallyvm.ExecuteTallyVm(tallyWasm, []string{"asdf"}, map[string]string{
		"VM_MODE":   "tally",
		"CONSENSUS": fmt.Sprintf("%v", true),
	})

	fmt.Println(vmRes)

	//rootCmd := cmd.NewRootCmd()
	//if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
	//      log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
	//      os.Exit(1)
	//}
}
