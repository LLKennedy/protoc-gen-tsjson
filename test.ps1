# Build and install the latest version of the plugin
go install .
# Builkd a test proto file
protoc --proto_path="." --tsjson_out="./tsout" ./root.proto