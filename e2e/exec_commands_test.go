package e2e

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	flagWasmType  = "wasm-type"
	flagTitle     = "title"
	flagSummary   = "summary"
	flagAuthority = "authority"
	flagDeposit   = "deposit"
	flagLabel     = "label"
	flagFixMsg    = "fix-msg"
	flagNoAdmin   = "no-admin"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyTxOptions(chainID string, options []flagOption) map[string]interface{} {
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

func applyQueryOptions(_ string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagOutput: "json",
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) executeQuery(ctx context.Context, c *chain, command []string, valIdx int, validation func([]byte, []byte) bool) {
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
		s.Require().FailNowf("Query execution validation failed", "stdout: %s, stderr: %s",
			string(stdOut), string(stdErr))
	}
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

func (s *IntegrationTestSuite) expectErrExecValidation(chain *chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut, stdErr []byte) bool {
		var txResp sdk.TxResponse
		gotErr := cdc.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			s.T().Logf("got tx error - stdout: %s", string(stdOut))
			s.T().Logf("got tx error - stderr: %s", string(stdErr))
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
	return func(stdOut, stdErr []byte) bool {
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
