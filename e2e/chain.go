package e2e

import (
	"fmt"
	"os"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/sedaprotocol/seda-chain/app"
)

const (
	keyringPassphrase = "testpassphrase"
	keyringAppName    = "testnet"
)

var (
	encodingConfig app.EncodingConfig
	cdc            codec.Codec
	txConfig       client.TxConfig
)

func init() {
	encodingConfig.Amino = codec.NewLegacyAmino()
	encodingConfig.InterfaceRegistry = types.NewInterfaceRegistry()
	encodingConfig.Marshaler = codec.NewProtoCodec(encodingConfig.InterfaceRegistry)
	encodingConfig.TxConfig = tx.NewTxConfig(encodingConfig.Marshaler, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	app.ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	cdc = encodingConfig.Marshaler
	txConfig = encodingConfig.TxConfig
}

type chain struct {
	dataDir         string
	id              string
	validators      []*validator
	genesisAccounts []*account // initial accounts in genesis
}

func newChain() (*chain, error) {
	tmpDir, err := os.MkdirTemp("", "e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &chain{
		id:      "chain-" + tmrand.Str(6),
		dataDir: tmpDir,
	}, nil
}

func (c *chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.dataDir, c.id)
}

func (c *chain) createAndInitValidators(count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-seda-%d", c.id, index),
	}
}
