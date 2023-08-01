# Developing

For setting up your environment to develop `seda-chain`. Shows how to build, run,
format, and clean the code base. To learn how to contribute please read
[here](CONTRIBUTING.md).

## Dev-Container

If you are using [VSCode](https://code.visualstudio.com/) and
[docker](https://www.docker.com/) you can open the project in a
[dev-container](https://github.com/Microsoft/vscode-remote-release) where all deps will be installed already.
Otherwise please see the [dev dependencies](#dev-dependencies).

## Dev Dependencies

### [Golang](https://go.dev/)

- [Golang](https://go.dev/dl/): you can download it from the linked page or:
    - Linux: Use your distribution's packagae manager.
    - Mac: Use `macports` or `brew`.
    - Windows: Use `scoop`

### [Ignite](https://docs.ignite.com/)

- Windows
  - Unfortunately ignite does not work for windows
- Mac & Linux:
  - Run the following command: `curl https://get.ignite.com/cli! | bash`

### [WasmVM](https://github.com/CosmWasm/wasmvm)

- `git clone https://github.com/CosmWasm/wasmvm`
- `make build-go`
- `make test`
- Make sure you put the shared library in your path by moving it to an appropiate location or adding it to your path.

- **NOTE**: This is currently not working on Windows.

## Building

**NOTE**: you must have the enviornment variable `CGO_ENABLED` set to `1`. This is becuase we use CosmWASM.

- `ignite chain build` will build and install the binary for you.
    - 
- `ignite chain serve` will build and run the chain without installing it.

## Formatting & Cleanliness

TODO

## Running

TODO

## Testing

TODO