package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *IntegrationTestSuite) testWasmStorageStoreOverlayWasm() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chain.id][0].GetHostPort("1317/tcp"))

	proposalCounter++
	proposalID := proposalCounter

	senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()

	s.execWasmStorageStoreOverlay(chainEndpoint, s.chain, 0, drWasm, "clean_title", "sustainable_summary", "data-request-executor", sender, standardFees.String(), false, proposalID)
	s.execGovVoteYes(chainEndpoint, s.chain, 0, sender, standardFees.String(), false, proposalID)

	// TO-DO test negative cases
	// - Overlay queries
	// - Proxy contract instantiation - query
	// - invalid wasm type
}

func (s *IntegrationTestSuite) execWasmStorageStoreOverlay(
	endpoint string,
	c *chain,
	valIdx int,
	overlayWasm,
	title,
	summary,
	wasmType,
	from,
	fees string,
	expectErr bool,
	proposalID int,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opt = append(opt, withKeyValue(flagWasmType, wasmType))
	opt = append(opt, withKeyValue(flagTitle, title))
	opt = append(opt, withKeyValue(flagSummary, summary))
	opt = append(opt, withKeyValue(flagDeposit, "10000000aseda"))
	opt = append(opt, withKeyValue(flagAuthority, authtypes.NewModuleAddress("gov").String()))

	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, overlayWasm)

	// 	$BIN tx wasm-storage submit-proposal store-overlay-wasm contract.wasm \
	//   --title nice_title --summary nice_summary
	//   --deposit 10000000aseda --authority $() \
	//   --wasm-type relayer --from $ADDR --keyring-backend test --gas auto --gas-adjustment 1.2 -y
	command := []string{
		binary,
		txCommand,
		types.ModuleName,
		"submit-proposal",
		"store-overlay-wasm",
		wasmFilePath,
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("proposing to store overlay Wasm %s on chain %s", wasmFilePath, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))

	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(endpoint, proposalID)
			s.Require().NoError(err)

			return proposal.GetProposal().Status == govtypesv1.StatusVotingPeriod
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) execGovVoteYes(
	endpoint string,
	c *chain,
	valIdx int,
	from,
	fees string,
	expectErr bool,
	proposalID int,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))

	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// $BIN tx gov vote 1 yes --from $ADDR --keyring-backend test --gas auto --gas-adjustment 1.2 -y
	command := []string{
		binary,
		txCommand,
		govtypes.ModuleName,
		"vote",
		strconv.Itoa(proposalID),
		"yes",
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("voting yes to proposal %s on chain %s", strconv.Itoa(proposalID), c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))

	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(endpoint, proposalID)
			s.Require().NoError(err)

			s.T().Logf("queried proposal %s  \n\n  got: %s\n\n", strconv.Itoa(proposalID), proposal.GetProposal().Status)

			return proposal.GetProposal().Status == govtypesv1.StatusPassed
		},
		30*time.Second,
		5*time.Second,
	)
}
