# Flow [![GoDoc](https://godoc.org/github.com/onflow/flow-go?status.svg)](https://godoc.org/github.com/onflow/flow-go)

Flow is a fast, secure, and developer-friendly blockchain built to support the next generation of games, apps and the
digital assets that power them. Read more about it [here](https://github.com/onflow/flow).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

## Table of Contents

- [Getting started](#getting-started)
- [Documentation](#documentation)
- [Installation](#installation)
    - [Clone Repository](#clone-repository)
    - [Install Dependencies](#install-dependencies)
- [Development Workflow](#development-workflow)
    - [Testing](#testing)
    - [Building](#building)
    - [Local Network](#local-network)
    - [Code Generation](#code-generation)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Getting started

- To install all dependencies and tools, see the [project setup](#installation) guide
- To dig into more documentation about Flow, see the [documentation](#documentation)
- To learn how to contribute, see the [contributing guide](/CONTRIBUTING.md)
- To see information on developing Flow, see the [development workflow](#development-workflow)

## Documentation

You can find an overview of the Flow architecture on the [documentation website](https://www.flow.com/primer).

Development on Flow is divided into work streams. Each work stream has a home directory containing high-level
documentation for the stream, as well as links to documentation for relevant components used by that work stream.

The following table lists all work streams and links to their home directory and documentation:

| Work Stream        | Home directory                             |
|--------------------|--------------------------------------------|
| Access Node        | [/cmd/access](/cmd/access)                 |
| Collection Node    | [/cmd/collection](/cmd/collection)         |
| Consensus Node     | [/cmd/consensus](/cmd/consensus)           |
| Execution Node     | [/cmd/execution](/cmd/execution)           |
| Verification Node  | [/cmd/verification](/cmd/verification)     |
| Observer Service   | [/cmd/observer](/cmd/observer)             |
| HotStuff           | [/consensus/hotstuff](/consensus/hotstuff) |
| Storage            | [/storage](/storage)                       |
| Ledger             | [/ledger](/ledger)                         |
| Networking         | [/network](/network/)                      |
| Cryptography       | [/crypto](/crypto)                         |

## Installation

- Clone this repository
- Install [Go](https://golang.org/doc/install) (Flow requires Go 1.23 and later)
- Install [Docker](https://docs.docker.com/get-docker/), which is used for running a local network and integration tests
- Make sure the [`GOPATH`](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable) and `GOBIN` environment variables
  are set, and `GOBIN` is added to your path:

    ```bash
    export GOPATH=$(go env GOPATH)
    export GOBIN=$GOPATH/bin
    export PATH=$PATH:$GOBIN
    ```

  Add these to your shell profile to persist them for future runs.
- Then, run the following command:

    ```bash
    make install-tools
    ```

At this point, you should be ready to build, test, and run Flow! 🎉

## Development Workflow

### Testing

Flow has a unit test suite and an integration test suite. Unit tests for a module live within the module they are
testing. Integration tests live in `integration/tests`.

Run the unit test suite:

```bash
make test
```

Run the integration test suite:

```bash
make integration-test
```

### Building

The recommended way to build and run Flow for local development is using Docker.

Build a Docker image for all nodes:

```bash
make docker-native-build-flow
```

Build a Docker image for a particular node role (replace `$ROLE` with `collection`, `consensus`, etc.):

```bash
make docker-native-build-$ROLE
```

#### Building a binary for the access node

Build the binary for an access node that can be run directly on the machine without using Docker.

```bash
make docker-native-build-access-binary
```
_this builds a binary for Linux/x86_64 machine_.

The make command will generate a binary called `flow_access_node`

### Importing the module

When importing the `github.com/onflow/flow-go` module in your Go project, testing or building your project may require setting extra Go flags because the module requires [cgo](https://pkg.go.dev/cmd/cgo). In particular, `CGO_ENABLED` must be set to `1` if `cgo` isn't enabled by default. This constraint comes from the underlying cryptography library. Refer to the [crypto repository build](https://github.com/onflow/crypto?tab=readme-ov-file#build) for more details.

### Local Network

A local version of the network can be run for manual testing and integration. See
the [Local Network Guide](/integration/localnet/README.md) for instructions.

### Code Generation

Generated code is kept up to date in the repository, so should be committed whenever it changes.

Run all code generators:

```bash
make generate
```

Generate protobuf stubs:

```bash
make generate-proto
```

Generate OpenAPI schema models:

```bash
make generate-openapi
```

Generate mocks used for unit tests:

```bash
make generate-mocks
```
