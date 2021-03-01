# protoc-gen-tsjson

Generate Typescript (and javascript) bindings for canonical JSON representation of proto messages, for when gRPC-web isn't doing it for you.

Primarily designed for use with [httpgrpc](https://github.com/LLKennedy/httpgrpc) but can be used on its own if you just want JSON proto message classes in your JS/TS project.

## Installing

TODO: confirm these instructions
```
go get github.com/LLKennedy/protoc-gen-tsjson
go install github.com/LLKennedy/protoc-gen-tsjson ./...
```

## Usage

```
protoc --proto_path=<paths> --tsjson_out=<output path> <proto files>
```