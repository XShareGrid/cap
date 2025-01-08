package ptk

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// 该文件提供一些webgrpc的框架工具函数

type ctxKey struct {
	Key string
}

var ctxValKeyUserInfo = &ctxKey{Key: "grpcUserInfo"}

const cookieHeader = "Cookie"

// MDSessionCookie 会话cookie
var MDSessionCookie = "SessionCookie"

// MDSessionIDKey md会话id
const MDSessionIDKey = "Authorization"

// MDRefererKey md referer
const MDRefererKey = "Referer"

// QueryUserID query key user id
const QueryUserID = "userID"

// GRPCRequestFilter grpc请求过滤器，
// 返回 rsp 回复
// 返回 err 错误
// 返回 handled 为true代表已处理，请求将直接返回，不再进行后续处理
type GRPCRequestFilter func(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) (handled bool, rsp interface{}, err error)
type grpcFilters struct {
	beforeHandler, afterHandler GRPCRequestFilter
}

var filters grpcFilters

// SetBeforeHandlerFilter 设置处理前的过滤器
func SetBeforeHandlerFilter(h GRPCRequestFilter) {
	filters.beforeHandler = h
}

// SetAfterHandlerFilter 设置处理后的过滤器
func SetAfterHandlerFilter(h GRPCRequestFilter) {
	filters.afterHandler = h
}

// GetRefererFromCtx 获取用户访问API的URL
func GetRefererFromCtx(ctx context.Context) (string, error) {
	// 添加用户信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", grpc.Errorf(codes.Internal, "failed to get metadata")

	}
	referer := md.Get(MDRefererKey)
	if len(referer) == 0 {
		return "", grpc.Errorf(codes.InvalidArgument, "no referer")
	}
	return referer[0], nil
}

var gServer *grpc.Server

// GraceStopWebGRPCServer 关闭服务
func GraceStopWebGRPCServer() {
	if gServer != nil {
		gServer.GracefulStop()
	}
}
