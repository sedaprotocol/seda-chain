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
	oracleProgram = "burner.wasm"
	tallyProgram  = "reflect.wasm"
	executorWasm  = "staking.wasm"
	coreWasm      = "core_contract.wasm"
)

func (s *IntegrationTestSuite) testWasmStorageStoreOracleProgram() {
	s.Run("store_oracle_programs", func() {
		senderAddress, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		sender := senderAddress.String()

		bytecode, err := os.ReadFile(filepath.Join(localWasmDirPath, oracleProgram))
		if err != nil {
			panic("failed to read an oracle program file")
		}
		opHash := crypto.Keccak256(bytecode)
		if opHash == nil {
			panic("failed to compute hash")
		}
		opHashHex := hex.EncodeToString(opHash)

		bytecode, err = os.ReadFile(filepath.Join(localWasmDirPath, tallyProgram))
		if err != nil {
			panic("failed to read a tally program file")
		}
		tallyHashBytes := crypto.Keccak256(bytecode)
		if tallyHashBytes == nil {
			panic("failed to compute hash")
		}
		tallyHashStr := hex.EncodeToString(tallyHashBytes)

		s.execWasmStorageStoreOracleProgram(s.chain, 0, oracleProgram, sender, standardFees.String(), false)
		s.execWasmStorageStoreOracleProgram(s.chain, 0, tallyProgram, sender, standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				// Query oracle programs individually.
				oracleRes, err := queryOracleProgram(s.endpoint, opHashHex)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(opHash, oracleRes.OracleProgram.Hash))
				tallyRes, err := queryOracleProgram(s.endpoint, tallyHashStr)
				s.Require().NoError(err)
				s.Require().True(bytes.Equal(tallyHashBytes, tallyRes.OracleProgram.Hash))

				// Query oracle programs at once.
				res, err := queryOraclePrograms(s.endpoint)
				s.Require().NoError(err)
				if len(res.List) == 2 {
					concat := strings.Join(res.List, "\n")
					strings.Contains(concat, opHashHex)
					strings.Contains(concat, tallyHashStr)
				}
				return true
			},
			30*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execWasmStorageStoreOracleProgram(
	c *chain,
	valIdx int,
	oracleProgram,
	from,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))

	opts := applyTxOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	wasmFilePath := filepath.Join(containerWasmDirPath, oracleProgram)

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

	s.T().Logf("storing an oracle program %s on chain %s", wasmFilePath, c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
