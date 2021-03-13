# protoc-gen-tsjson

Generate Typescript (and javascript) bindings for canonical JSON representation of proto messages, for when gRPC-web isn't doing it for you.

Primarily designed for use with [httpgrpc](https://github.com/LLKennedy/httpgrpc) but can be used on its own if you just want JSON proto message classes in your JS/TS project.

## Pre-requisites

If you cannot obtain a pre-built binary (e.g. the releases section of this project, once it has some) you will require Go >= 1.16 to compile the protoc plugin.

## Installing

```
go install github.com/LLKennedy/protoc-gen-tsjson
```

Alternatively, building the binary for your environment ahead of time and ensuring the result is on the path will also work. Protoc simply adds "protoc-gen" to any "plugin_out"-style options then attempts to execute "protoc-gen-plugin" blindly. Although Go was used to write this plugin, and is required to build the binary, it is not required for use with protoc, or for use with the output files.

## Usage

```
protoc --proto_path=<paths> --tsjson_out=<output path> <proto files>
```