# Contributing

This file describes the process for contributing to `seda-chain`.


## Starting

First and foremost, [fork](https://github.com/sedaprotocol/seda-chain/fork) the
repository. Then please read the [developing instructions](DEVELOPING.md) for 
setting up your environment.


## Commits

Your commits must follow specific guidelines.

### Signed Commits

Sign all commits with a GPG key. GitHub has extensive documentation on how to:

- [Create](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key)
  a new GPG key.
- [Add](https://docs.github.com/en/authentication/managing-commit-signature-verification/adding-a-gpg-key-to-your-github-account)
  a GPG key to your GitHub.
- [Sign](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
  your commits.

### Conventional Commits

All commits are to follow the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) standard.
Commit messages should always be meaningful.


## Error Handling

If you already have an error you can return, just return it. If you would like 
to add context to the error, use `errorsmod.Wrap()` or `errorsmod.Wrapf()`. 
For instance:

```go
import (
	errorsmod "cosmossdk.io/errors"
)

err := cdc.UnmarshalJSON(bz, &data)
if err != nil {
	return errorsmod.Wrapf(err, "failed to unmarshal %s genesis state", types.ModuleName)
}
```

If there is no error at hand, you need to create an error or make use of SDK 
errors. To create a new error, add an error in `types/errors.go` and return it. 
To make use of SDK errors, find an appropriate error from Cosmos SDK package's 
`types/errors/errors.go` and return it. For example, we return a `ErrInvalidRequest` 
error when there is a failure in preliminary validation of transaction request 
parameters.

```go
import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)
func (m *MsgCreateVestingAccount) ValidateBasic() error {
	if m.EndTime <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}
	return nil
}
```


## Formatting and Cleanliness

Please run `make lint` before making a commit to format the code.


## Creating a Pull Request

For creating the PR, please follow the instructions below.

1. Firstly, please open a
   [PR](https://github.com/sedaprotocol/seda-chain/compare) from your forked repo
   to the `main` branch of `seda-chain`.
2. Please fill in the PR template that is there.
3. Then assign it to yourself and anyone else who worked on the issue with you.
4. Make sure all CI tests pass.
5. Finally, please assign at least two reviewers to your PR:
   - [FranklinWaller](https://github.com/FranklinWaller)
   - [gluax](https://github.com/gluax)
   - [mariocao](https://github.com/mariocao)
   - [Thomasvdam](https://github.com/Thomasvdam)
   - [hacheigriega](https://github.com/hacheigriega)
   - [NikolaCehic95](https://github.com/NikolaCehic95)
