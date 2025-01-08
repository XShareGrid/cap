package i18n

import (
	"context"

	"github.com/XShareGrid/cap/logger"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
)

// GetUserLanguageByGrpc 从上下文中提取 Accept-Language，返回一个 languageStr
func GetUserLanguageByMeta(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	//把md的内容 加到logger里面
	//logger.CAP.Info("metadata:", md)
	if !ok {
		logger.CAP.Info("Context does not contain metadata")
		return EN
	}
	languages_value := md.Get("Accept-Language")
	if len(languages_value) == 0 {
		logger.CAP.Info("Accept-Language metadata is missing")
		//return EN
		return ZH
	}

	userLanguage := languages_value[0]
	switch userLanguage {
	case defaultUserLanguageZhCN:
		userLanguage = ZH
	case defaultUserLanguageEnUS:
		userLanguage = EN
	}

	return userLanguage
}

// SaveUserLanguageToCtx 将用户语言ctx保存到ctx中
func SaveUserLanguageToCtx(ctx context.Context, grpcCtx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(grpcCtx)
	if ok {
		ctx = metadata.NewIncomingContext(ctx, md)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

func WithUserLanguage(ctx context.Context, langCode string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	logger.CAP.Info("metadata:", md)
	if !ok {
		logger.CAP.Info("Context does not contain metadata")
		return ctx
	}
	md.Set("Accept-Language", langCode)
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}

// GetLanguageType 从GRPC ctx获取语言类型
func GetLanguageTypeByMeta(ctx context.Context) language.Tag {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return language.Make(defaultUserLanguageEnUS)
	}
	languages_value := md.Get("Accept-Language")
	if len(languages_value) == 0 {
		//return language.Make(defaultUserLanguageEnUS)
		return language.Make(defaultUserLanguageZhCN)
	}

	language_str := languages_value[0]

	return language.Make(language_str)
}

func NewOutgoingCtxWithUserLanguage(ctx context.Context, userLanguage string) context.Context {
	acceptLang := defaultUserLanguageEnUS
	switch userLanguage {
	case ZH, defaultUserLanguageZhCN:
		acceptLang = defaultUserLanguageZhCN
	case EN, defaultUserLanguageEnUS:
		acceptLang = defaultUserLanguageEnUS
	}
	md := metadata.Pairs("accept-language", acceptLang)
	return metadata.NewOutgoingContext(ctx, md)
}

func NewOutgoingCtxWithGRPCCtx(ctx context.Context, grpcCtx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(grpcCtx)
	if ok {
		ctx = metadata.NewIncomingContext(ctx, md)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
