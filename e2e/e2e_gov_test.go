package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/hyperledger/burrow/crypto"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *IntegrationTestSuite) testWasmStorageStoreOverlayWasm() {
	proposalCounter++
	proposalID := proposalCounter

	senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddress.String()

	bytecode, err := os.ReadFile(filepath.Join(localWasmDirPath, overlayWasm))
	if err != nil {
		panic("failed to read data request Wasm file")
	}
	overlayHashBytes := crypto.Keccak256(bytecode)
	if overlayHashBytes == nil {
		panic("failed to compute hash")
	}
	overlayHashStr := hex.EncodeToString(overlayHashBytes)

	s.execWasmStorageStoreOverlay(s.chain, 0, overlayWasm, "clean_title", "sustainable_summary", "data-request-executor", sender, standardFees.String(), false, proposalID)
	s.execGovVoteYes(s.chain, 0, sender, standardFees.String(), false, proposalID)

	s.Require().Eventually(
		func() bool {
			overlayWasmRes, err := queryOverlayWasm(s.endpoint, overlayHashStr)
			s.Require().NoError(err)
			s.Require().True(bytes.Equal(overlayHashBytes, overlayWasmRes.Wasm.Hash))

			wasms, err := queryOverlayWasms(s.endpoint)
			s.Require().NoError(err)

			return fmt.Sprintf("%s,%s", overlayHashStr, types.WasmTypeDataRequestExecutor.String()) == wasms.HashTypePairs[0]
		},
		30*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) execWasmStorageStoreOverlay(
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
			proposal, err := queryGovProposal(s.endpoint, proposalID)
			s.Require().NoError(err)

			return proposal.GetProposal().Status == govtypesv1.StatusVotingPeriod
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) execGovVoteYes(
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
			proposal, err := queryGovProposal(s.endpoint, proposalID)
			s.Require().NoError(err)

			return proposal.GetProposal().Status == govtypesv1.StatusPassed
		},
		30*time.Second,
		5*time.Second,
	)
}