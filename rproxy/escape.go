package rproxy

import "strings"

var slashEscapeChar = "__slash__"

// 因为beego的bug https://github.com/astaxie/beego/issues/3942
// 对url中带%2f斜杠的参数进行手动转义
// 前端对 / 转义为 %2f
// EAP将 %2f 转义为 __slash__  ==> SlashEscape
// 子系统将 __slash__ 转义为 / ==> SlashUnEscape

// SlashEscape 斜杠转义
func SlashEscape(url string) string {
	return strings.Replace(
		strings.Replace(url, "%2f", slashEscapeChar, -1),
		"%2F", slashEscapeChar, -1)
}

// SlashUnEscape 斜杠反转义
func SlashUnEscape(url string) string {
	return strings.Replace(url, slashEscapeChar, "/", -1)

}
