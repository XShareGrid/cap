# msg/errors

errors包定义了基于msg封装的UserError，实现了error的接口
- 支持更丰富的错误记录
- 堆栈调用
- 多国语言

## 定义多语言错误

查看[i18n](../i18n/README.md)模块

## 使用示例  

查看[example](example/main.go)


## 与原生错误集成

### 接口说明

- **Wrap**: 将任意error转为一个UserError

```go
ue := errors.Wrap(fmt.Errorf("错误1"))
ue.Log()  // => 错误1
```
- **UserError.Triggers**: 使用错误，触发一个新的错误，UserError将会记录错误的历史信息

```go
ue := errors.New("usererror")
newErr := error.Wrap(fmt.Errorf("错误1")).Triggers(fmt.Errorf("错误2")).Triggers(ue)
newErr.DumpErrors().Log() // => usererror, 错误2, 错误1
```

- **Match**: 匹配两个错误
- **MatchAll**: 匹配两个错误（包括历史关联错误）

```go
u0 := fmt.Errorf("u0")
u1 := errors.Wrap(u0)
u2 := u1.Triggers(errors.New("u2"))
fmt.Println(errors.Match(u0, u1))     // => true
fmt.Println(errors.Match(u1, u2))     // => false
fmt.Println(errors.MatchAll(u1, u2))  // => true
```

- **UserError.FillDebugArgs**: 填充调试的参数，一般用于填充预定义错误中的占位符
- **UserError.FillIDAndArgs**: 填充用户错误的ID与参数，一般用于直接在下层返回的错误中添加用户错误信息
```go
// 调用下层接口返回一个错误
err := CallXXXXX()
// 直接包装这个错误，并且添加用户错误信息
ue := errors.Wrap(err).FillIDAndArgs(ue.ERR_SESSION_EXPIRED)
fmt.Println(err.TrError())   // => 会话过期，请重新登陆 <nil>
```

### 示例

查看[example](example/wrap.go)


## 预定义错误
为了使错误可判，一般会将包里的错误做预定义，使用New接口或PreDef接口进行预定义错误：

```go
var Err1 = errors.New("some error")
var Err2 = errors.PreDef("some error")
```

系统会为预定义的错误分配uint类型的uid，以方便快速比较