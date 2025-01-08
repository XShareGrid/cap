package template

import "github.com/XShareGrid/cap/msg/errors"

// ErrOperatePermissionDenied 没有权限操作其他人创建的模板
var ErrOperatePermissionDenied = errors.New("no access to make an operation to the template created by other user")

// XSG-3148
// ErrDisableSharingFormToAllUsers 禁止向所有用户共享表单
var ErrDisableSharingFormToAllUsers = errors.New("Disable sharing form to all users")
