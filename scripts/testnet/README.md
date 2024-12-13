## Testnet Deployment Scripts

These scripts generate a genesis file and deploy chain across nodes following the given configuration.

First, make sure to create a configuration file named `config.sh` using the template `config_example.sh`.
Then, check out the tag as specified by the configuration's `CHAIN_VERSION` and run `make build`.
Finally, run the scripts in the following order to generate the genesis file and deploy the chain:

1. `create_genesis.sh` - Creates a genesis file with given parameters and token allocations.
2. `add_groups_to_genesis.sh` - Adds groups to the genesis file (optional).
3. `upload_and_start.sh` - Uploads and runs `setup_node.sh` on the nodes. Then uploads the validator files and starts the nodes.
