# Indexing Plugin

A simple plugin to aid with indexing the blockchain state. It listens to state changes, decodes the keys/values, and publishes these changes in JSON format to an SNS topic.

This plugin follows the architecture laid out in [ADR-038](https://github.com/cosmos/cosmos-sdk/blob/main/docs/architecture/adr-038-state-listening.md).

## Usage

Currently Cosmos SDK only supports 1 plugin names `ABCI`, so that's what we're using. After building the plugin take note of the location of the executable as you'll need to provide that in the environment variables.

The process that starts the node should have the following environment variables set.

```sh
export COSMOS_SDK_ABCI=PATH_TO_PLUGIN_EXECUTABLE
export SNS_TOPIC_ARN=""
export S3_LARGE_MSG_BUCKET_NAME=""
export PLUGIN_LOG_FILE=PATH_TO_DESIRED_LOG_FILE
# Optionally you can also specify the log level, one of "trace", "debug", "info", "warn", "error"
export PLUGIN_LOG_LEVEL="WARN"
# Optionally you can specifiy a comma separated list of event types which are allowed to be published on the queue.
# When omitted it will default to publishing all message types.
export ALLOWED_MESSAGE_TYPES="block,account"
```

Lastly, as we're using SNS and S3 the node needs access to a valid set of AWS credentials with permission to publish messages to the specified topic and upload access to the specified bucket.

### Logging

The plugin uses [Hashicorp's hclog](https://pkg.go.dev/github.com/hashicorp/go-hclog) as this is the recommended approach in go-plugins and is also what the StreamingManager sets up on the Cosmos SDK side.

Since we've been unable to get the plugin log output to show up in the node output for now we resort to logging to a file. A benefit is that this makes monitoring the plugin easier since we don't have to filter out any node logs. When deploying the plugin and configuring the log file make sure to set up log rotation for the log output directory.

### Disaster Recovery

When the plugin fails to process updates it halts the node (provided the `stop-node-on-err` setting in the `app.toml` file is set to `true). This makes it easier for us to resume the indexing process from the point where the failure occurred, but there are a few caveats to keep in mind with how the node handles these errors.

If we simplify things there are 3 points of failure in the plugin:

1. [Errors during the initialisation of the plugin](#plugin-initialisation-errors).
2. [Errors while processing the `ListenFinalizeBlock` handler](#listenfinalizeblock-errors).
3. [Errors while processing the `ListenCommit` handler](#listencommit-errors).

#### Plugin Initialisation Errors

Most of these errors should be related to incorrect configuration or missing environment variables. These errors will prevent the SEDA node process from starting altogether with a trace like the following:

```txt
panic: failed to load streaming plugin: <MESSAGE_FROM_PLUGIN>
This usually means
  the plugin was not compiled for this architecture,
  the plugin is missing dynamic-link libraries necessary to run,
  the plugin is not executable by this process due to file permissions, or
  the plugin failed to negotiate the initial go-plugin protocol handshake

Additional notes about plugin:
  Path: /bin/sh
  Mode: -rwxr-xr-x
  Owner: 0 [root] (current: 501 [user])
  Group: 0 [wheel] (current: 20 [staff])

# Omitted long stacktrace which is not very helpful.
```

`<MESSAGE_FROM_PLUGIN>` usually is a string with the error message from the plugin, but it could be missing in case it's an unhandled panic thrown somewhere during the plugin initialisation.

Since the node didn't even start running we don't have to worry about block/state mismatches and can just fix the plugin/configuration and restart the process.

#### ListenFinalizeBlock Errors

These errors are are most likely related to the processing of block/transaction data or publishing the messages on the queue. These errors halt the chain with an error message like the following:

```txt
1:16PM ERR FinalizeBlock listening hook failed err="rpc error: code = Unknown desc = <MESSAGE_FROM_PLUGIN>" height=XXX module=server
2024-03-26T13:16:10.507+0100 [INFO]  plugin.abci: plugin process exited: plugin=/bin/sh id=14714
```

`<MESSAGE_FROM_PLUGIN>` usually is a string with the error message from the plugin, but it could be missing in case it's an unhandled error/panic.

These errors are also fairly easy to recover from. When restarting the node it will 'notice' a discrepancy between the state height and store height, where the store height (N) is where the error occurred:

```txt
1:16PM INF ABCI Replay Blocks appHeight=N-1 module=consensus stateHeight=N-1 storeHeight=N
1:16PM INF Replay last block using real app module=consensus
```

As the logs indicate the node will replay the last block in order to update the state and app. This means the `ListenFinalizeBlock` handler is called again for height N. As long as the indexer is capabale of handling duplicate messages there are no further actions to take. Provided that the plugin has been fixed/problem has been resolved the indexing should be able to continue.

#### ListenCommit Errors

These errors are are most likely related to the processing of state data or publishing the messages on the queue. These errors halt the chain with an error message like the following:

```txt
1:12PM ERR Commit listening hook failed err="rpc error: code = Unknown desc = <MESSAGE_FROM_PLUGIN>" height=404 module=server
2024-03-26T13:12:24.905+0100 [INFO]  plugin.abci: plugin process exited: plugin=/bin/sh id=13031
```

`<MESSAGE_FROM_PLUGIN>` usually is a string with the error message from the plugin, but it could be missing in case it's an unhandled error/panic.

These errors are the most tricky to recover from as the node will have updated its state and app, so restarting will resume the node at height N+1, where N is the height at which the error occurred. In order to reprocess height N the node needs to be rolled back 1 block:

```sh
sedad rollback
# Rolled back state to height N and hash XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX%
```

> [!WARNING]
> Even though this command does not invoke any callbacks on the plugin it still initialises it and as such requires the required environment variables to be present in the shell that executes the command.

Now the node can be restarted and it should resume from height N. It will call both `ListenFinalizeBlock` and `ListenCommit` again for that height, so as long as the indexer is capable of handling duplicate messages everything should be able to continue (provided that the plugin has been fixed/problem has been resolved).

## Building

```sh
go build -o PATH_TO_PLUGIN_EXECUTABLE ./plugins/indexing/plugin.go
# Alternatively, outputs in the /build directory in the project root
make build-plugin
```

## Local Development

To simplify local development we use [a SNS emulator](https://github.com/Admiral-Piett/goaws/) and [a S3 emulator](https://github.com/adobe/S3Mock). To connect to this from the plugin you need to need to build the plugin with the `dev` flag. In addition you'll need specifiy an environment variable for `SNS_ENDPOINT` in the process that launches the node, and an environment variable for `S3_ENDPOINT` (which should correspond to your local port of the service.).

```sh
# Example urls
export SNS_TOPIC_ARN="arn:aws:sns:eu-west-2:queue:local-updates-topic"
export SNS_ENDPOINT=http://localhost:4100
export S3_LARGE_MSG_BUCKET_NAME="indexer-localnet-large-messages"
export S3_ENDPOINT=http://localhost:9444
```

```sh
go build --tags dev -o PATH_TO_PLUGIN_EXECUTABLE ./plugins/indexing/plugin.go
# Alternatively, outputs in the /build directory in the project root
make build-plugin-dev
```
