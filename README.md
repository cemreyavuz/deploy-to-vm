# deploy-to-vm

[![codecov](https://codecov.io/github/cemreyavuz/deploy-to-vm/graph/badge.svg?token=0XCAR85Q87)](https://codecov.io/github/cemreyavuz/deploy-to-vm)

`deploy-to-vm` is an application used for deploying web applications to virtual machines.

## Development

### Installing dependencies

```sh
go mod tidy
```

### Running the application

```sh
go run ./...
```

### Building the application

```sh
go build -o deploy-to-vm.bin ./cmd/deploy-to-vm
```

### Testing the application

```sh
go test ./...
```
