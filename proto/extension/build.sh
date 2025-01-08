protoc -I. -I$GOOGLE_PROTO -I$EH_PROJECT_PATH/energy-backend/webproto --go_opt=paths=source_relative --go_out=plugins=grpc:./ *.proto
