# PROTOC VERSION 29.3
# https://github.com/protocolbuffers/protobuf/releases/tag/v29.3
export CAP_PROTOC_PATH=/usr/local/bin/protoc293

export PROTOC=$CAP_PROTOC_PATH/bin/protoc


# libprotoc 29.3
$PROTOC --version

  
$PROTOC -I. -I$CAP_PROTOC_PATH/include -I. --go_opt=paths=source_relative --go_out=./ enum.proto
