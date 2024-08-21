# Developing

For setting up your environment to develop `seda-chain`. Shows how to build, run,
format, and clean the code base. To learn how to contribute please read
[here](CONTRIBUTING.md).

> [!NOTE]
> Windows is not supported at this time.

## Dev-Container

If you are using [VSCode](https://code.visualstudio.com/) and
[docker](https://www.docker.com/) you can open the project in a
[dev-container](https://github.com/Microsoft/vscode-remote-release) where all deps will be installed already.
Otherwise please see the [dev dependencies](#dev-dependencies).

## Dev Dependencies

### [clang-format](https://clang.llvm.org/docs/ClangFormat.html)

We use clang format to format our protobuf generated code.

- Linux
  - Please see your distro specific installation tool(i.e `apt`) and use that to install it.
- Macos:
  - Using [brew](https://brew.sh/): `brew install clang-format`
  - Using [macports](https://www.macports.org/): `sudo port install clang-format`

### [docker](https://www.docker.com/)

Docker is used to help make release and static builds locally.

- Linux
  - Please see your distro specific installation tool(i.e `apt`) and use that to install it.
- Macos:
  - Using [brew](https://brew.sh/): `brew install --cask docker`
  - Using [macports](https://www.macports.org/): `sudo port install docker`

### [Golang](https://go.dev/)

We use Golang as the language to develop `sedad` as it has the [CosmosSDK](https://v1.cosmos.network/sdk).

- [Golang](https://go.dev/dl/): you can download it from the linked page or:
  - Linux: Use your distribution's package manager.
  - Mac: Use `macports` or `brew`.
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

A necessary tool for generating protobuf code.

- Linux:
  - Please see your distro specific installation tool(i.e `apt`) and use that to install it.
- Macos:
  - Using [brew](https://brew.sh/): `brew install protobuf`
  - Using [macports](https://www.macports.org/): `sudo port install protobuf-cpp`

#### Protobuf Sub-Deps

We also need some dependencies to make protobuf work for cosmos.

##### Buf

The `buf` tool.

- Linux:
  - `curl -sSL https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64 -o buf && chmod +x buf && sudo mv buf /usr/local/bin`
- Macos:
  - Using [brew](https://brew.sh/): `brew install bufbuild/buf/buf`
  - Using [macports](https://www.macports.org/): `sudo port install buf`

### [WasmVM](https://github.com/CosmWasm/wasmvm)

WasmVM is the library that makes CosmWASM possible.

You can install that by running:

```bash
sudo ./scripts/install_wasmvm.sh
```

## Building Using Make

To build the protobuf (only necessary if you made changes in the proto files) you will need to run:

```bash
make prot-dep-install
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
BIN=./build/sedad

$BIN tendermint unsafe-reset-all
rm -rf ~/.sedad

$BIN init node0 --default-denom aseda --chain-id seda-1-local

$BIN keys add satoshi --keyring-backend test
$BIN add-genesis-account $($BIN keys show satoshi --keyring-backend test -a) 10000000000000000seda
$BIN gentx satoshi 10000000000000000seda --keyring-backend test --chain-id seda-1-local
$BIN collect-gentxs
$BIN start
```

## Linting & Formatting

To lint and format the protobuf(only necessary if you mess with protobuf):

```bash
make proto-fmt
make proto-lint
```

If you have not install a Go linters runner, install it first:

```bash
make lint-install
```

Run format and run linter for go sided:

```bash
make fmt
make lint
```

## Running

After running the `make install` command you should be able to use `sedad --help`.

## Testing

To run all unit tests:

```bash
make test-unit
```

To see test coverage:

```bash
make cover-html
```

To run end-to-end tests:

```bash
GITHUB_TOKEN=<your_github_pat> make test-e2e
```
