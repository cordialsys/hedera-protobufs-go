This fork contains alternative method for generating go grpc
files from hedera-protobufs repo.

# Hedera™ Hashgraph Go Protubufs

> Generated protobufs in Go for interacting with Hedera Hashgraph: the official distributed
> consensus platform built using the hashgraph consensus algorithm for fast,
> fair and secure transactions. Hedera enables and empowers developers to
> build an entirely new class of decentralized applications.

## Install

```sh
$ go get github.com/cordialsys/hedera-protobufs-go
```

## Usage

```go
import "github.com/cordialsys/hedera-protobufs-go/services"
```

## Development

When updating the protobufs submodule, the generated code should be updated.

### Prerequisites

-   [Go](https://golang.org/doc/install) v25+

-   [Protobuf Compiler](https://developers.google.com/protocol-buffers)

-   Go plugins for the protobuf compiler

    ```sh
    $ go install google.golang.org/protobuf/cmd/protoc-gen-go
    $ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
    ```

### Build

```sh
# generate go code from protobuf definitions
$ go run ./internal/cmd/build

# ensure the projects build
$ go vet ./...
```

## License

Licensed under Apache License,
Version 2.0 – see [LICENSE](LICENSE) in this repo
