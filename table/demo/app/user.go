package main

import (
	"context"

	cap "github.com/XShareGrid/cap/table/proto/go"
)

type TestUserInfoProvider struct {
}

func (*TestUserInfoProvider) GetCurrentUser(ctx context.Context) (*cap.UserInfo, error) {
	return &cap.UserInfo{
		Id:          1,
		UserName:    "demo",
		DisplayName: "演示",
	}, nil
}
func (*TestUserInfoProvider) GetUserInfoByID(id int32) (*cap.UserInfo, error) {
	return &cap.UserInfo{
		Id:          1,
		UserName:    "demo",
		DisplayName: "演示",
	}, nil
}
