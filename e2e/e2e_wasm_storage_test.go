package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	drWasm       = "burner.wasm"
	tallyWasm    = "reflect.wasm"
	executorWasm = "staking.wasm"
	coreWasm     = "core_contract.wasm"
)

func (s *IntegrationTestSuite) testWasmStorageStoreOracleProgram() {
	s.Run("store_an_oracle_program", func() {
		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		bytecode, err := os.ReadFile(filepath.Join(localWasmDirPath, drWasm))
		if err != nil {
			panic("failed to read a wasm file")
		}
		drHashBytes := crypto.Keccak256(bytecode)
		if drHashBytes == nil {
			panic("failed to compute hash")
		}
		drHashStr := hex.EncodeToString(drHashBytes)

		//
		bytecode, err = os.ReadFile(filepath.Join(localWasmDirPath, tallyWasm))
		if err != nil {
			panic("failed to read tally Wasm file")
		}
		tallyHashBytes := crypto.Keccak256(bytecode)
		if tallyHashBytes == nil {
			panic("failed to compute hash")
		}
		tallyHashStr := hex.EncodeToString(tallyHashBytes)

		s.execWasmStorageStoreDataRequest(s.chain, 0, drWasm, "data-request", sender, standardFees.String(), false)
		s.execWasmStorageStoreDataRequest(s.chain, 0, tallyWasm, "tally", sender, standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				// Query wasms individually.
				drWasmRes, err := queryOracleProgram(s.endpoint, drHashStr)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(drHashBytes, drWasmRes.Wasm.Hash))
				tallyWasmRes, err := queryOracleProgram(s.endpoint, tallyHashStr)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(tallyHashBytes, tallyWasmRes.Wasm.Hash))

				// Query wasms at once.
				res, err := queryOraclePrograms(s.endpoint)
				s.Require().NoError(err)
				if len(res.List) == 2 {
					concat := strings.Join(res.List, "\n")
					strings.Contains(concat, drHashStr)
					strings.Contains(concat, tallyHashStr)
				}
				return false
			},
			30*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execWasmStorageStoreDataRequest(
	c *chain,
	valIdx int,
	drWasm,
	wasmType,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opt = append(opt, withKeyValue(flagWasmType, wasmType))

	opts := applyTxOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, drWasm)

	command := []string{
		binary,
		txCommand,
		types.ModuleName,
		"store-oracle-program",
		wasmFilePath,
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("storing a wasm file %s on chain %s", wasmFilePath, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
