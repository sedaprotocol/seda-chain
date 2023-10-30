//nolint:unused
package e2e

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	flagFrom            = "from"
	flagHome            = "home"
	flagFees            = "fees"
	flagGas             = "gas"
	flagOutput          = "output"
	flagChainID         = "chain-id"
	flagSpendLimit      = "spend-limit"
	flagGasAdjustment   = "gas-adjustment"
	flagFeeAccount      = "fee-account"
	flagBroadcastMode   = "broadcast-mode"
	flagKeyringBackend  = "keyring-backend"
	flagAllowedMessages = "allowed-messages"

	// wasm-storage flags
	flagWasmType = "wasm-type"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(chainID string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagGasAdjustment:  "1.5",
		flagChainID:        chainID,
		flagHome:           containerChainHomePath,
		flagFees:           standardFees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) execEncode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing seda-chain encoding with %v", c.id, txPath)
	command := []string{
		binary,
		txCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	s.executeTx(ctx, c, command, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully encode with %v", txPath)
	return encoded
}

func (s *IntegrationTestSuite) execDecode(
	c *chain,
	txPath string,
	opt ...flagOption,
) string {
	opts := applyOptions(c.id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing seda-chain decoding with %v", c.id, txPath)
	command := []string{
		binary,
		txCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	s.executeTx(ctx, c, command, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	s.T().Logf("successfully decode %v", txPath)
	return decoded
}

type txBankSend struct {
	from      string
	to        string
	amt       string
	fees      string
	log       string
	expectErr bool
}

func (s *IntegrationTestSuite) execBankSendBatch(
	c *chain,
	valIdx int, //nolint:unparam
	txs ...txBankSend,
) int {
	sucessBankSendCount := 0

	for i := range txs {
		s.T().Logf(txs[i].log)

		s.execBankSend(c, valIdx, txs[i].from, txs[i].to, txs[i].amt, txs[i].fees, txs[i].expectErr)
		if !txs[i].expectErr {
			if !txs[i].expectErr {
				sucessBankSendCount++
			}
		}
	}

	return sucessBankSendCount
}

func (s *IntegrationTestSuite) runGovExec(c *chain, valIdx int, submitterAddr, govCommand string, proposalFlags []string, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	command := []string{
		binary,
		txCommand,
		govtypes.ModuleName,
		govCommand,
	}

	generalFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "300000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	command = concatFlags(command, proposalFlags, generalFlags)

	s.T().Logf("Executing seda-chain tx gov %s on chain %s", govCommand, c.id)
	s.executeTx(ctx, c, command, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Logf("Successfully executed %s", govCommand)
}

func (s *IntegrationTestSuite) executeTx(ctx context.Context, c *chain, command []string, valIdx int, validation func([]byte, []byte) bool) {
	s.T().Logf("executing command %s", command)

	if validation == nil {
		validation = s.defaultExecValidation(s.chain, 0)
	}
	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)
	exec, err := s.dkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    s.valResources[c.id][valIdx].Container.ID,
		User:         "root",
		Cmd:          command,
	})
	s.Require().NoError(err)

	err = s.dkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	s.Require().NoError(err)

	stdOut := outBuf.Bytes()
	stdErr := errBuf.Bytes()

	if !validation(stdOut, stdErr) {
		s.Require().FailNowf("Exec validation failed", "stdout: %s, stderr: %s",
			string(stdOut), string(stdErr))
	}
}

// TO-DO refactor
func (s *IntegrationTestSuite) expectErrExecValidation(chain *chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		gotErr := cdc.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			fmt.Println("stdout: " + string(stdOut))
			s.Require().True(expectErr)
		}

		endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))

		// wait for the tx to be committed on chain
		s.Require().Eventuallyf(
			func() bool {
				gotErr := queryTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s",
			string(stdOut), string(stdErr),
		)
		return true
	}
}

func (s *IntegrationTestSuite) defaultExecValidation(chain *chain, valIdx int) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp sdk.TxResponse
		if err := cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.String(), "code: 0") || txResp.Code == 0 {
			endpoint := fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
			s.Require().Eventually(
				func() bool {
					return queryTx(endpoint, txResp.TxHash) == nil
				},
				time.Minute,
				5*time.Second,
				"stdOut: %s, stdErr: %s",
				string(stdOut), string(stdErr),
			)
			return true
		}
		return false
	}
}
