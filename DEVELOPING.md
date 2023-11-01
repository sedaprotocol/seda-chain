# Developing

For setting up your environment to develop `seda-chain`. Shows how to build, run,
format, and clean the code base. To learn how to contribute please read
[here](CONTRIBUTING.md).

**NOTE**: Windows is not supported at this time.

## Dev-Container

If you are using [VSCode](https://code.visualstudio.com/) and
[docker](https://www.docker.com/) you can open the project in a
[dev-container](https://github.com/Microsoft/vscode-remote-release) where all deps will be installed already.
Otherwise please see the [dev dependencies](#dev-dependencies).

## Dev Dependencies

### [clang-format]()

### [Golang](https://go.dev/)
We use Golang as the language to develop `seda-chaind` as it has the [CosmosSDK](https://v1.cosmos.network/sdk).

- [Golang](https://go.dev/dl/): you can download it from the linked page or:
    - Linux: Use your distribution's packagae manager.
    - Mac: Use `macports` or `brew`.
    - Windows: Use `scoop`
- Ensure that `$GOPATH` and `$PATH` have been set properly. On a Mac that uses the Z shell, you may have to run the following:
```zsh
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.zprofile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.zprofile
echo "export GO111MODULE=on" >> ~/.zprofile
source ~/.zprofile
```
### [make](https://www.gnu.org/software/make/)

We use GNU Make to help us built, lint, fmt, and etc for our project.

- Linux:
  - Your distro likely already comes with this. You can check by running `which make`.
  - Otherwise please see your distro specific installation tool(i.e `apt`) and use that to install it.
- Macos:
  - You can check if it's already installed by `which make`.
  - Otherwise use [brew](https://brew.sh/) or [macports](https://www.macports.org/) to install it.

<!-- It actually uses docker to run protobuf commmands... this should be fixed -->
### [Protobuf](https://protobuf.dev/)

- Linux:
  - Please see your distro specific installation tool(i.e `apt`) and use that to install it.
- Macos:
  - Using [brew](https://brew.sh/): `brew install protobuf`
  - Using [macports](https://www.macports.org/): `sudo port install protobuf-cpp`

#### Protobuf Sub-Deps


<!-- TODO this needs to be tested more -->
### [WasmVM](https://github.com/CosmWasm/wasmvm)

- `git clone https://github.com/CosmWasm/wasmvm`
- `make build-go`
- `make test`
- Make sure you put the shared library in your path by moving it to an appropriate location or adding it to your path.

- **NOTE**: This is currently not working on Windows.

## Building using Make

To build the protobuf(only necessary if you change the protobuf) you will need to run,:
```bash
make proto-update-deps
make proto-gen
```

To build, run:
```bash
make build
```

To install (builds and moves the executable to `$GOPATH/bin`, which should be in `$PATH`), run:
```bash
make install
```

## Running a Single-node Local Testnet

To run a single-node testnet locally:
```bash
make build
BIN=./build/seda-chaind

$BIN tendermint unsafe-reset-all
rm -rf ~/.seda-chain
$BIN init new node0

$BIN keys add satoshi --keyring-backend test
$BIN add-genesis-account $($BIN keys show satoshi --keyring-backend test -a) 10000000000000000seda
$BIN gentx satoshi 10000000000000000seda --keyring-backend test
$BIN collect-gentxs
$BIN start
```

## Linting

<!-- @gluax add clang-format as a dep -->
<!-- @gluax this cmd doesn't work... -->
To lint the protobuf(only necessary if you mess with protobuf):
```bash
make proto-lint
```

If you have not install a Go linters runner, install it first:
```bash
make lint-install
```

Run linters with auto-fix
```bash
make lint-fix
```

or without auto-fix:
```bash
make lint
```

## Running

After running the `make install` command you should be able to use `seda-chaind --help`.

## Testing

To test the node so far we only have unit tests:
```bash
make test-unit
```