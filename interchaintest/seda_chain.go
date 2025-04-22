package interchaintest

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const valKey = "validator"

// SEDAChain wraps cosmos.CosmosChain to provide custom logic.
type SEDAChain struct {
	*cosmos.CosmosChain
	log *zap.Logger
}

func NewSEDAChain(cosmosChain *cosmos.CosmosChain, logger *zap.Logger) *SEDAChain {
	return &SEDAChain{
		CosmosChain: cosmosChain,
		log:         logger,
	}
}

func (c *SEDAChain) Config() ibc.ChainConfig {
	return c.CosmosChain.Config()
}

// Bootstraps the chain and starts it from genesis
func (c *SEDAChain) Start(testName string, ctx context.Context, additionalGenesisWallets ...ibc.WalletAmount) error {
	chainCfg := c.Config()

	decimalPow := int64(math.Pow10(int(*chainCfg.CoinDecimals)))

	genesisAmount := types.Coin{
		Amount: sdkmath.NewInt(10_000_000).MulRaw(decimalPow),
		Denom:  chainCfg.Denom,
	}

	genesisSelfDelegation := types.Coin{
		Amount: sdkmath.NewInt(5_000_000).MulRaw(decimalPow),
		Denom:  chainCfg.Denom,
	}

	if chainCfg.ModifyGenesisAmounts != nil {
		genesisAmount, genesisSelfDelegation = chainCfg.ModifyGenesisAmounts()
	}

	genesisAmounts := []types.Coin{genesisAmount}

	configFileOverrides := chainCfg.ConfigFileOverrides

	eg := new(errgroup.Group)
	// Initialize config and sign gentx for each validator.
	for _, v := range c.Validators {
		v := v
		v.Validator = true
		eg.Go(func() error {
			if err := v.InitFullNodeFiles(ctx); err != nil {
				return err
			}
			for configFile, modifiedConfig := range configFileOverrides {
				modifiedToml, ok := modifiedConfig.(testutil.Toml)
				if !ok {
					return fmt.Errorf("Provided toml override for file %s is of type (%T). Expected (DecodedToml)", configFile, modifiedConfig)
				}
				if err := testutil.ModifyTomlConfigFile(
					ctx,
					c.log,
					v.DockerClient,
					v.TestName,
					v.VolumeName,
					configFile,
					modifiedToml,
				); err != nil {
					return fmt.Errorf("failed to modify toml config file: %w", err)
				}
			}
			if !c.Config().SkipGenTx {
				return initValidatorGenTx(ctx, v, genesisAmounts, genesisSelfDelegation)
			}
			return nil
		})
	}

	// Initialize config for each full node.
	for _, n := range c.FullNodes {
		n := n
		n.Validator = false
		eg.Go(func() error {
			if err := n.InitFullNodeFiles(ctx); err != nil {
				return err
			}
			for configFile, modifiedConfig := range configFileOverrides {
				modifiedToml, ok := modifiedConfig.(testutil.Toml)
				if !ok {
					return fmt.Errorf("Provided toml override for file %s is of type (%T). Expected (DecodedToml)", configFile, modifiedConfig)
				}
				if err := testutil.ModifyTomlConfigFile(
					ctx,
					c.log,
					n.DockerClient,
					n.TestName,
					n.VolumeName,
					configFile,
					modifiedToml,
				); err != nil {
					return err
				}
			}
			return nil
		})
	}

	// wait for this to finish
	if err := eg.Wait(); err != nil {
		return err
	}

	if c.Config().PreGenesis != nil {
		err := c.Config().PreGenesis(chainCfg)
		if err != nil {
			return err
		}
	}

	// for the validators we need to collect the gentxs and the accounts
	// to the first node's genesis file
	validator0 := c.Validators[0]
	for i := 1; i < len(c.Validators); i++ {
		validatorN := c.Validators[i]

		bech32, err := validatorN.AccountKeyBech32(ctx, valKey)
		if err != nil {
			return err
		}

		if err := validator0.AddGenesisAccount(ctx, bech32, genesisAmounts); err != nil {
			return err
		}

		if !c.Config().SkipGenTx {
			if err := copyGentx(ctx, validatorN, validator0); err != nil {
				return err
			}
		}
	}

	for _, wallet := range additionalGenesisWallets {

		if err := validator0.AddGenesisAccount(ctx, wallet.Address, []types.Coin{{Denom: wallet.Denom, Amount: wallet.Amount}}); err != nil {
			return err
		}
	}

	if !c.Config().SkipGenTx {
		if err := validator0.CollectGentxs(ctx); err != nil {
			return err
		}
	}

	genbz, err := validator0.GenesisFileContent(ctx)
	if err != nil {
		return err
	}

	genbz = bytes.ReplaceAll(genbz, []byte(`"stake"`), []byte(fmt.Sprintf(`"%s"`, chainCfg.Denom)))

	if c.Config().ModifyGenesis != nil {
		genbz, err = c.Config().ModifyGenesis(chainCfg, genbz)
		if err != nil {
			return err
		}
	}

	// Provide EXPORT_GENESIS_FILE_PATH and EXPORT_GENESIS_CHAIN to help debug genesis file
	exportGenesis := os.Getenv("EXPORT_GENESIS_FILE_PATH")
	exportGenesisChain := os.Getenv("EXPORT_GENESIS_CHAIN")
	if exportGenesis != "" && exportGenesisChain == c.Config().Name {
		c.log.Debug("Exporting genesis file",
			zap.String("chain", exportGenesisChain),
			zap.String("path", exportGenesis),
		)
		_ = os.WriteFile(exportGenesis, genbz, 0600)
	}

	chainNodes := c.Nodes()

	for _, cn := range chainNodes {
		if err := cn.OverwriteGenesisFile(ctx, genbz); err != nil {
			return err
		}
	}

	if err := chainNodes.LogGenesisHashes(ctx); err != nil {
		return err
	}

	// Sidecar processes are not supported for SEDA Chain.
	// eg, egCtx := errgroup.WithContext(ctx)
	// for _, s := range c.Sidecars {
	// 	s := s
	// 	err = s.containerLifecycle.Running(ctx)
	// 	if s.preStart && err != nil {
	// 		eg.Go(func() error {
	// 			if err := s.CreateContainer(egCtx); err != nil {
	// 				return err
	// 			}
	// 			if err := s.StartContainer(egCtx); err != nil {
	// 				return err
	// 			}
	// 			return nil
	// 		})
	// 	}
	// }
	// if err := eg.Wait(); err != nil {
	// 	return err
	// }

	eg, egCtx := errgroup.WithContext(ctx)
	for _, n := range chainNodes {
		n := n
		eg.Go(func() error {
			return n.CreateNodeContainer(egCtx)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	peers := chainNodes.PeerString(ctx)

	eg, egCtx = errgroup.WithContext(ctx)
	for _, n := range chainNodes {
		n := n
		c.log.Info("Starting container", zap.String("container", n.Name()))
		eg.Go(func() error {
			if err := n.SetPeers(egCtx, peers); err != nil {
				return err
			}
			return n.StartContainer(egCtx)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	// Wait for blocks before considering the chains "started"
	return testutil.WaitForBlocks(ctx, 2, c.GetNode())
}

func copyGentx(ctx context.Context, src, dst *cosmos.ChainNode) error {
	nid, err := src.NodeID(ctx)
	if err != nil {
		return fmt.Errorf("getting node ID: %w", err)
	}

	relPath := fmt.Sprintf("config/gentx/gentx-%s.json", nid)

	gentx, err := src.ReadFile(ctx, relPath)
	if err != nil {
		return fmt.Errorf("getting gentx content: %w", err)
	}

	err = dst.WriteFile(ctx, gentx, relPath)
	if err != nil {
		return fmt.Errorf("overwriting gentx: %w", err)
	}

	return nil
}

func initValidatorGenTx(ctx context.Context, tn *cosmos.ChainNode, genesisAmounts []types.Coin, genesisSelfDelegation types.Coin) error {
	if err := tn.CreateKey(ctx, valKey); err != nil {
		return err
	}
	bech32, err := tn.AccountKeyBech32(ctx, valKey)
	if err != nil {
		return err
	}
	if err := tn.AddGenesisAccount(ctx, bech32, genesisAmounts); err != nil {
		return err
	}

	// tn.lock.Lock()
	// defer tn.lock.Unlock()

	var command []string
	command = append(command, "gentx", valKey, fmt.Sprintf("%s%s", genesisSelfDelegation.Amount.String(), genesisSelfDelegation.Denom),
		"--keyring-backend", keyring.BackendTest,
		"--chain-id", tn.Chain.Config().ChainID,
		"--key-file-no-encryption",
	)
	_, _, err = tn.ExecBin(ctx, command...)
	return err
}
