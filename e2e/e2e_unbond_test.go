package e2e

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *IntegrationTestSuite) testUnbond() {
	s.Run("unbond", func() {
		val1Addr, err := s.chain.validators[0].keyInfo.GetAddress()
		s.Require().NoError(err)
		val1 := val1Addr.String()

		val2Addr, err := s.chain.validators[1].keyInfo.GetAddress()
		s.Require().NoError(err)
		val2 := val2Addr.String()

		s.execUnbond(s.chain, 0, sdk.ValAddress(val1Addr), val1, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, 2)
				if err != nil {
					return false
				}
				return len(batch.BatchSignatures) == len(s.chain.validators)
			},
			30*time.Second,
			5*time.Second,
		)

		s.execUnbond(s.chain, 1, sdk.ValAddress(val2Addr), val2, standardFees.String(), false)
		s.Require().Eventually(
			func() bool {
				batch, err := queryBatch(s.endpoint, 3)
				if err != nil {
					return false
				}

				if len(batch.BatchSignatures) != len(s.chain.validators) {
					return false
				}
				for _, sig := range batch.BatchSignatures {
					if len(sig.Secp256K1Signature) != 65 {
						return false
					}
					if sig.ValidatorAddress.Empty() {
						return false
					}
				}
				return true
			},
			90*time.Second,
			5*time.Second,
		)
	})
}

func (s *IntegrationTestSuite) execUnbond(
	c *chain,
	valIdx int,
	valAddr sdktypes.ValAddress,
	from string,
	fees string,
	expectErr bool,
	opt ...flagOption,
) {
	opt = append(opt, withKeyValue(flagFees, fees))
	opt = append(opt, withKeyValue(flagFrom, from))
	opts := applyTxOptions(c.id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	command := []string{
		binary,
		txCommand,
		stakingtypes.ModuleName,
		"unbond",
		valAddr.String(),
		"1aseda",
		"-y",
	}
	for flag, value := range opts {
		command = append(command, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("unbond tx to trigger batch creation on chain %s", c.id)

	s.executeTx(ctx, c, command, valIdx, s.expectErrExecValidation(c, valIdx, expectErr))
}
