# PROTOC VERSION 29.3
# https://github.com/protocolbuffers/protobuf/releases/tag/v29.3
export CAP_PROTOC_PATH=/usr/local/bin/protoc293

export PROTOC=$CAP_PROTOC_PATH/bin/protoc


# libprotoc 29.3
# protoc-gen-go v1.36.2
# protoc-gen-go-grpc 1.5.1
# protoc-gen-grpc-gateway version:
# Version v2.25.1, commit unknown, built at unknown
# protoc-gen-openapiv2 version:
# Version v2.25.1, commit unknown, built at unknown
$PROTOC --version
protoc-gen-go --version
protoc-gen-go-grpc --version
echo "protoc-gen-grpc-gateway version:"
protoc-gen-grpc-gateway --version
echo "protoc-gen-openapiv2 version:"
protoc-gen-openapiv2 --version
$PROTOC -I. -I../../proto/vendor -I$CAP_PROTOC_PATH/include -I../../proto --go_out=./go --go_opt=paths=source_relative \
--go-grpc_out=./go --go-grpc_opt=paths=source_relative --go-grpc_opt=require_unimplemented_servers=false \
--grpc-gateway_out=./go --grpc-gateway_opt=paths=source_relative \
table_wservice.proto

  
