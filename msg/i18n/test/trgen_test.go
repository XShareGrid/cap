package test

import (
	"testing"

	"github.com/XShareGrid/cap/msg/i18n"
)

func TestGenerateTranslatorCode(t *testing.T) {
	i18n.GenerateTranslatorCode()

}

// func TestTranslate(t *testing.T) {
// 	fmt.Println(i18n.Translate(Lang_enUS, ERR_1))
// 	fmt.Println(i18n.Translate(Lang_zhCN, ERR_1))
// 	fmt.Println(i18n.Translate(Lang_enUS, ERR_2))
// 	fmt.Println(i18n.Translate(Lang_zhCN, ERR_2))
// }

// func TestLanguage(t *testing.T) {
// 	matcher := language.NewMatcher([]language.Tag{Lang_enUS, Lang_zhCN})
// 	tag, _, c := matcher.Match(language.SimplifiedChinese)
// 	fmt.Println(tag)
// }
