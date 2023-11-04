package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *IntegrationTestSuite) testInstantiateAndRegisterProxyContract() {
	proposalCounter++
	proposalID := proposalCounter

	senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()

	_, err = os.ReadFile(filepath.Join(localWasmDirPath, proxyWasm))
	s.Require().NoError(err)

	s.execWasmStore(s.chain, 0, proxyWasm, sender, standardFees.String(), false)
	s.execInstantiateAndRegisterProxyContract(s.chain, 0, "clean_title", "sustainable_summary", "data-request-executor", sender, standardFees.String(), false, proposalID)
	s.execGovVoteYes(s.chain, 0, sender, standardFees.String(), false, proposalID)

	s.Require().Eventually(
		func() bool {
			res, err := queryProxyContractRegistry(s.endpoint)
			s.Require().NoError(err)

			_, err = sdktypes.AccAddressFromBech32(res.Address)
			s.Require().NoError(err)

			return res.Address != ""
		},
		30*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) execInstantiateAndRegisterProxyContract(
	c *chain,
	valIdx int,
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

	opt = append(opt, withKeyValue(flagNoAdmin, "true"))
	opt = append(opt, withKeyValue(flagFixMsg, "true"))

	opt = append(opt, withKeyValue(flagDeposit, "10000000aseda"))
	opt = append(opt, withKeyValue(flagTitle, title))
	opt = append(opt, withKeyValue(flagSummary, summary))
	opt = append(opt, withKeyValue(flagLabel, "fortunate_label"))

	opt = append(opt, withKeyValue(flagAuthority, authtypes.NewModuleAddress("gov").String()))

	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	codeID := "1"
	command := []string{
		binary,
		txCommand,
		types.ModuleName,
		"submit-proposal",
		"instantiate-and-register-proxy-contract",
		codeID,
		"{\"token\":\"aseda\"}",
		"74657374696e67", // salt
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("proposing to instantiate and register as proxy contract (code ID %s) on chain %s", codeID, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))

	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(s.endpoint, proposalID)
			s.Require().NoError(err)

			return proposal.GetProposal().Status == govtypesv1.StatusVotingPeriod
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) execWasmStore(
	c *chain,
	valIdx int,
	drWasm,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, drWasm)

	command := []string{
		binary,
		txCommand,
		wasmtypes.ModuleName,
		"store",
		wasmFilePath,
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("storing data request Wasm %s on chain %s", wasmFilePath, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
