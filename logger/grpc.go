package logger

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

func grpcLogInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error) {
	if info.FullMethod != "/psi.Platform/GetLoginUser" {
		GRPC.Debug("request received, method[%s], req[%+v]", info.FullMethod, req)
	}
	rsp, err = handler(ctx, req)

	if info.FullMethod != "/psi.Platform/GetLoginUser" && info.FullMethod != "/psi.Subsystem/GetInfo" {
		content := rsp.(proto.Message).String()
		if len(content) > 100 {
			content = content[0:100]
			content += "..."
		}
		GRPC.Debug("request handled, method[%s], rsp[%v], err[%+v]", info.FullMethod, content, err)
	}
	return
}

func GRPCLogger() grpc.ServerOption {
	return grpc.UnaryInterceptor(grpcLogInterceptor)
}
