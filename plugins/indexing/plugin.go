package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	streamingabci "cosmossdk.io/store/streaming/abci"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/sedaprotocol/seda-chain/app"
	"github.com/sedaprotocol/seda-chain/app/params"

	authmodule "github.com/sedaprotocol/seda-chain/plugins/indexing/auth"
	bankmodule "github.com/sedaprotocol/seda-chain/plugins/indexing/bank"
	base "github.com/sedaprotocol/seda-chain/plugins/indexing/base"
	log "github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	pluginsqs "github.com/sedaprotocol/seda-chain/plugins/indexing/sqs"
	types "github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

var _ storetypes.ABCIListener = &IndexerPlugin{}

// IndexerPlugin is the implementation of the baseapp.ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type IndexerPlugin struct {
	block     *types.BlockContext
	cdc       codec.Codec
	sqsClient *pluginsqs.SqsClient
	logger    *log.Logger
}

func (p *IndexerPlugin) publishToQueue(messages []*types.Message) error {
	publishError := p.sqsClient.PublishToQueue(messages)
	if publishError != nil {
		p.logger.Error("Failed to publish messages to queue.", "error", publishError)
		return publishError
	}

	return nil
}

func (p *IndexerPlugin) ListenFinalizeBlock(_ context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	p.logger.Debug(fmt.Sprintf("[%d] Start processing finalize block.", req.Height))
	p.block = types.NewBlockContext(req.Height, req.Time)

	// TODO(#229) Change to +2 to account for the votes message
	messages := make([]*types.Message, 0, len(req.Txs)+1)

	blockMessage, err := base.ExtractBlockUpdate(p.block, req)
	if err != nil {
		p.logger.Error("Failed to extract block update", "error", err)
		return err
	}
	messages = append(messages, blockMessage)

	txMessages, err := base.ExtractTransactionUpdates(p.block, p.cdc, req, res)
	if err != nil {
		p.logger.Error("Failed to extract Tx updates", "error", err)
		return err
	}
	messages = append(messages, txMessages...)
	// TODO(#229) Extract all vote data.

	if err := p.publishToQueue(messages); err != nil {
		return err
	}

	p.logger.Debug(fmt.Sprintf("[%d] Processed finalize block.", req.Height))
	return nil
}

func (p *IndexerPlugin) extractUpdate(change *storetypes.StoreKVPair) (*types.Message, error) {
	switch change.StoreKey {
	case bankmodule.StoreKey:
		return bankmodule.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case authmodule.StoreKey:
		return authmodule.ExtractUpdate(p.block, p.cdc, p.logger, change)
	default:
		return nil, nil
	}
}

func (p *IndexerPlugin) ListenCommit(_ context.Context, _ abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	p.logger.Debug(fmt.Sprintf("[%d] Start processing commit", p.block.Height))
	var messages []*types.Message

	for _, change := range changeSet {
		message, err := p.extractUpdate(change)
		if err != nil {
			p.logger.Error("Failed to extract update", "error", err)
			return err
		}

		if message != nil {
			p.logger.Debug("Extracted update", "message", message)
			messages = append(messages, message)
		}
	}

	if err := p.publishToQueue(messages); err != nil {
		return err
	}

	p.logger.Debug(fmt.Sprintf("[%d] Processed commit", p.block.Height))
	return nil
}

func main() {
	logFile, err := log.GetLogFile()
	if err != nil {
		// Printing the error makes it easier to see what went wrong in the plugin output
		fmt.Println(err)
		panic(err)
	}
	defer logFile.Close()

	logger := log.NewLogger(logFile)
	logger.Info("initialising plugin")

	// Configure address prefixes.
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	cfg.Seal()

	// Prepare Proto codec.
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		logger.Fatal("failed to create interface registry", err)
	}
	std.RegisterInterfaces(interfaceRegistry)
	app.ModuleBasics.RegisterInterfaces(interfaceRegistry)

	sqsClient, err := pluginsqs.NewSqsClient()
	if err != nil {
		logger.Fatal("failed to create sqs client", err)
	}

	filePlugin := &IndexerPlugin{
		cdc:       codec.NewProtoCodec(interfaceRegistry),
		sqsClient: sqsClient,
		logger:    logger,
	}

	logger.Info("finished initialising plugin")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: streamingabci.Handshake,
		Plugins: map[string]plugin.Plugin{
			"abci": &streamingabci.ListenerGRPCPlugin{Impl: filePlugin},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
