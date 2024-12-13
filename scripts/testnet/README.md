## Testnet Deployment Scripts

These scripts generate a genesis file and deploy chain across the nodes specified in the parameters.

Make sure to first create a configuration file named config.sh using the template config_example.sh and populate it with the values reflecting your environment and desired deployment settings.

Run in the following order, one by one:

1. `create_genesis.sh` - Validates validator files for each node and creates a genesis file.
2. `build_genesis_state.sh` - Runs a chain, deploys Wasm contracts on it, and dumps the Wasm state, which is then added to a given genesis file. Must be used with at least one of the following flags: `--add-groups` and `add-wasm-contracts`.
3. `upload_and_start.sh` - Uploads and runs setup_node.sh on the nodes to process necessary setups. Then uploads the validator files and genesis and restarts the nodes.
