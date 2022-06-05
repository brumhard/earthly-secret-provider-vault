# `earthly-secret-provider-vault`

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)

An earthly secret provider for Hashicorp Vault.

The project uses `earthly` to make your life easier. If you're not familiar with earthly you can take a look at [the docs](https://docs.earthly.dev/).

## Setup

- [install earthly](https://earthly.dev/get-earthly)
- run `earthly +local-setup` to setup the `.githooks` directory

## Test & lint

Run linting

```bash
earthly +lint
```

Run tests

```bash
earthly +test
```
