# msg
msg包定义了一种通用的消息msg.Msg，可以用为错误、日志的统一格式（基类）
## 字段及含义
字段|说明
--|--
Timestamp | 时间戳
ID | 消息ID
Args | 消息参数
Severity | 优先级
Custom | 用户自定义字段

## 接口

- **Msg::GetMessage**  

获取翻译后的消息


