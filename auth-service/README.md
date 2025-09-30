## generate stub
export GOOGLEAPIS_PATH=$(go list -m -f '{{ .Dir }}' github.com/googleapis/googleapis)

protoc -I . \
-I "$GOOGLEAPIS_PATH" \
--go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--descriptor_set_out=./envoy/app.desc \
--include_imports \
--include_source_info \
stub/*.proto

