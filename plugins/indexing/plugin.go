package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-plugin"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/gogoproto/proto"

	streamingabci "cosmossdk.io/store/streaming/abci"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"

	circuittypes "cosmossdk.io/x/circuit/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcfee "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v8/modules/core/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
	vestingtypes "github.com/sedaprotocol/seda-chain/x/vesting/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/auth"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/bank"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/base"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/batching"
	dataproxy "github.com/sedaprotocol/seda-chain/plugins/indexing/data-proxy"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	oracleprogram "github.com/sedaprotocol/seda-chain/plugins/indexing/oracle-program"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/pluginaws"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/pubkey"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/tally"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

var _ storetypes.ABCIListener = &IndexerPlugin{}

// IndexerPlugin is the implementation of the baseapp.ABCIListener interface
// For Go plugins this is all that is required to process data sent over gRPC.
type IndexerPlugin struct {
	block     *types.BlockContext
	cdc       codec.Codec
	sqsClient *pluginaws.SqsClient
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

	p.logger.Trace(fmt.Sprintf("[%d] Extracting block update.", req.Height))
	blockMessage, err := base.ExtractBlockUpdate(p.block, req, res)
	if err != nil {
		p.logger.Error("[ListenFinalizeBlock] Failed to extract block update", "error", err)
		return err
	}
	messages = append(messages, blockMessage)

	p.logger.Trace(fmt.Sprintf("[%d] Extracting transaction updates.", req.Height))
	txMessages, err := base.ExtractTransactionUpdates(p.block, p.cdc, p.logger, req, res)
	if err != nil {
		p.logger.Error("[ListenFinalizeBlock] Failed to extract Tx updates", "error", err)
		return err
	}
	messages = append(messages, txMessages...)
	// TODO(#229) Extract all vote data.

	p.logger.Trace(fmt.Sprintf("[%d] Publishing messages to queue.", req.Height))
	if err := p.publishToQueue(messages); err != nil {
		p.logger.Error("[ListenFinalizeBlock] Failed to publish messages to queue", "error", err)
		return err
	}

	p.logger.Debug(fmt.Sprintf("[%d] Processed finalize block.", req.Height))
	return nil
}

func (p *IndexerPlugin) extractUpdate(change *storetypes.StoreKVPair) (*types.Message, error) {
	switch change.StoreKey {
	case bank.StoreKey:
		return bank.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case auth.StoreKey:
		return auth.ExtractUpdate(p.block, p.cdc, p.logger, change)
	// Enable when indexer supports these messages
	// case staking.StoreKey:
	// 	return staking.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case pubkey.StoreKey:
		return pubkey.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case dataproxy.StoreKey:
		return dataproxy.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case batching.StoreKey:
		return batching.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case oracleprogram.StoreKey:
		return oracleprogram.ExtractUpdate(p.block, p.cdc, p.logger, change)
	case tally.StoreKey:
		return tally.ExtractUpdate(p.block, p.cdc, p.logger, change)
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
			p.logger.Error("[ListenCommit] Failed to extract update", "error", err)
			return err
		}

		if message != nil {
			p.logger.Debug("Extracted update", "message", message)
			messages = append(messages, message)
		}
	}

	if err := p.publishToQueue(messages); err != nil {
		p.logger.Error("[ListenCommit] Failed to publish messages to queue", "error", err)
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
	// Rather than relying on the app module we manually register the interfaces for all modules,
	// as this significantly reduces the size of the binary.
	// genutil doesn't have any interfaces to register
	authtypes.RegisterInterfaces(interfaceRegistry)
	authz.RegisterInterfaces(interfaceRegistry)
	vestingtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	feegrant.RegisterInterfaces(interfaceRegistry)
	group.RegisterInterfaces(interfaceRegistry)
	govtypesv1.RegisterInterfaces(interfaceRegistry)
	govtypesv1beta1.RegisterInterfaces(interfaceRegistry)
	minttypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	upgradetypes.RegisterInterfaces(interfaceRegistry)
	evidencetypes.RegisterInterfaces(interfaceRegistry)
	consensustypes.RegisterInterfaces(interfaceRegistry)
	circuittypes.RegisterInterfaces(interfaceRegistry)
	// capability doesn't have any interfaces to register
	wasmtypes.RegisterInterfaces(interfaceRegistry)
	ibctypes.RegisterInterfaces(interfaceRegistry)
	ibctm.RegisterInterfaces(interfaceRegistry)
	ibcfee.RegisterInterfaces(interfaceRegistry)
	ibctransfertypes.RegisterInterfaces(interfaceRegistry)
	pubkeytypes.RegisterInterfaces(interfaceRegistry)
	icatypes.RegisterInterfaces(interfaceRegistry)
	crisistypes.RegisterInterfaces(interfaceRegistry)
	packetforwardtypes.RegisterInterfaces(interfaceRegistry)
	wasmstoragetypes.RegisterInterfaces(interfaceRegistry)
	tallytypes.RegisterInterfaces(interfaceRegistry)
	dataproxytypes.RegisterInterfaces(interfaceRegistry)
	batchingtypes.RegisterInterfaces(interfaceRegistry)

	sqsClient, err := pluginaws.NewSqsClient(logger)
	if err != nil {
		logger.Fatal("failed to create AWS clients", err)
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
