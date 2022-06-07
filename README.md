# `earthly-secret-provider-vault`

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)

An earthly secret provider for Hashicorp Vault using the currently experimental [`secret_provider` config field](https://docs.earthly.dev/docs/earthly-config#secret_provider-experimental).

## Sneak Peek

```Earthfile
some-target:
	RUN --secret VAULT_SECRET=+secrets/some/path.some_field \
		echo "secret from vault at path 'some/path' with field 'some_field': $VAULT_SECRET"
```

## Installation

### From source

If you have Go 1.16+, you can directly install by running:

```bash
go install github.com/brumhard/earthly-secret-provider-vault/cmd/earthly-secret-provider-vault@latest
```

> Based on your go configuration the binary can be found in `$GOPATH/bin` or `$HOME/go/bin` in case `$GOPATH` is not set.
> Make sure to add the respective directory to your `$PATH`.  
> [For more information see go docs for further information](https://golang.org/ref/mod#go-install). Run `go env` to view your current configuration.

### From the released binaries

Download the desired version for your operating system and processor architecture from the [releases page](https://github.com/brumhard/earthly-secret-provider-vault/releases).
Make the file executable and place it in a directory available in your `$PATH`.

### Docker

The secret provider is also distributed as docker images.

- `ghcr.io/brumhard/earthly-secret-provider-vault:vX.X.X` as a docker image based on `distroless:static` containing only the binary
- `ghcr.io/brumhard/earthly-secret-provider-vault:vX.X.X-full` based on the upstream `earthly/earthly` image and already configured to use this secret provider. This can be used as a drop-in replacement for the earthly image in CI. Be aware that you still need to set vault specific configs (see setup)

## Usage

### Setup

The only things necessary to use the secret provider are installing it (see [Installation](#installation) section) and setting the earthly config by:

```shell
earthly config global.secret_provider "earthly-secret-provider-vault"
```

Since the secret providers can't use any environment variables you also need to set some vault specific configs.
To not reimplement all authentication logic defined in the `vault` CLI the secret provider only uses a token that can be generated with the `vault` CLI.
This could look sth like the following for `userpass` authentication.

```shell
export VAULT_ADDR=<your-vault-addr>
vault login --method=userpass username=someone
earthly-secret-provider-vault config address $VAULT_ADDR
earthly-secret-provider-vault config token $(vault print token)
```

### Using vault secrets in earthly

An example Earthfile using the secret provider looks like the following:

```Earthfile
VERSION 0.6
FROM golang:1.18-alpine

test:
    RUN --no-cache --secret TEST=+secrets/path/to/secret.field \
        echo "top secret $TEST"
```

Be aware that the `+secrets/` prefix is normally used for the cloud secret provider by earthly. If the secret is not found in the vault it will also be looked up in the earthly cloud.

The syntax to access a secret in vault is `<vault-path>/<vault-subpath>.<field>`.
The lookup that would happen in this example could be replicated with the following vault CLI command:

```shell
vault kv get -field field path/to/secret
```

### Configuration

All configuration options can be set either via the `config` subcommand as described above or in the configuration file that is placed at `~/.earthly/vault.yml` by default. That is also the file that is edited by the `config` subcommand.

The following configuration options are available:

| Name      | Description                                                                                                                                                                                                                                                                                                                                                                                                 | Required |
| :-------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :------: |
| `address` | The address of the vault server.                                                                                                                                                                                                                                                                                                                                                                            |    x     |
| `token`   | The token to use for authentication. This can be generated with the vault CLI for example. `earthly-secret-provider-vault` will also try to read the token from `~/.vault-token` which is used by the vault CLI to store a token after login. This is only used if the token is not set in the `vault.yml`.                                                                                                 |    x     |
| `prefix`  | A prefix that is prepended to all paths that are looked up with the secret provider. For example if all your secrets are at `root/cicd/` you can use that as prefix and only define the rest of the path in the Earthfile. E.g. if your full path is `root/cicd/some_app/config.field` and the prefix is `root/cicd` you can use the secret in earthly with `--secret TEST=+secrets/some_app/config.field`. |          |

## CLI

`earthly-secret-provider-vault` provides a little CLI on top of the secret provider functionality.
This can be used to print the version and set config options.
For further information have a look at the [CLI docs](docs/earthly-secret-provider-vault.md) or run `earthly-secret-provider-vault --help`.
