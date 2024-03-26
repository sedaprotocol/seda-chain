# Indexing Plugin

A simple plugin to aid with indexing the blockchain state. It listens to state changes, decodes the keys/values, and publishes these changes in JSON format on an SQS queue.

This plugin follows the architecture laid out in [ADR-038](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-038-state-listening.md).

## Usage

Currently Cosmos SDK only supports 1 plugin names `ABCI`, so that's what we're using. After building the plugin take note of the location of the executable as you'll need to provide that in the environment variables.

The process that starts the node should have the following environment variables set.

```sh
export COSMOS_SDK_ABCI=PATH_TO_PLUGIN_EXECUTABLE
export SQS_QUEUE_URL=""
export PLUGIN_LOG_FILE=PATH_TO_DESIRED_LOG_FILE
# Optionally you can also specify the log level, one of "trace", "debug", "info", "warn", "error"
export PLUGIN_LOG_LEVEL="WARN"
```

Lastly, as we're using SQS the node needs access to a valid set of AWS credentials with permission to publish messages to the specified queue.

### Logging

The plugin uses [Hashicorp's hclog](https://pkg.go.dev/github.com/hashicorp/go-hclog) as this is the recommended approach in go-plugins and is also what the StreamingManager sets up on the Cosmos SDK side.

Since we've been unable to get the plugin log output to show up in the node output for now we resort to logging to a file. A benefit is that this makes monitoring the plugin easier since we don't have to filter out any node logs. When deploying the plugin and configuring the log file make sure to set up log rotation for the log output directory.

## Building

```sh
go build -o PATH_TO_PLUGIN_EXECUTABLE ./plugins/indexing/plugin.go
```

## Local Development

To simplify local development we use [a SQS emulator](https://github.com/Admiral-Piett/goaws/). To connect to this from the plugin you need to need to build the plugin with the `dev` flag. In addition you'll need specifiy an environment variable for `SQS_ENDPOINT` (which should be the base of the `SQS_QUEUE_URL`) in the process that launches the node.

```sh
# Example urls
export SQS_QUEUE_URL=http://localhost/4100/test-queue.fifo
export SQS_ENDPOINT=http://localhost:4100
```

```sh
go build --tags dev -o PATH_TO_PLUGIN_EXECUTABLE ./plugins/indexing/plugin.go
```
