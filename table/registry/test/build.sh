protoc -I. -I$GOOGLE_PROTO -I$GOPATH/src --go_out=plugins=grpc:./ enum.proto
