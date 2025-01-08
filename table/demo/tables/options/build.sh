protoc -I. -I$GOOGLE_PROTO -I$EH_PROJECT_PATH/webproto --go_opt=paths=source_relative --go_out=plugins=grpc:./ *.proto
