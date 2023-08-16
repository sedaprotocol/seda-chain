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
- Ensure that `$GOPATH` and `$PATH` have been set properly. On a Mac, you may have to run the following:
```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.zprofile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.zprofile
echo "export GO111MODULE=on" >> ~/.zprofile
source ~/.zprofile
```
 
### [Ignite (Optional)](https://docs.ignite.com/)

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

## Building using Make

To build, run:
```bash
$ make build
```

To install (builds and moves the executable to `$GOPATH/bin`, which should be in `$PATH`), run:
```bash
$ make install
```

## Building Using Ignite

**NOTE**: you must have the enviornment variable `CGO_ENABLED` set to `1`. This is becuase we use CosmWASM.

- `ignite chain build` will build and install the binary for you.
    - 
- `ignite chain serve` will build and run the chain without installing it.

## Linting

If you have not install a Go linters runner, install it first:
```bash
$ make lint-install
```

Run linters with auto-fix
```bash
$ make lint-fix
```

or without auto-fix:
```bash
$ make lint
```

## Running

TODO

## Testing

TODO