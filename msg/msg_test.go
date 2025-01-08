package msg

import (
	"fmt"
	"gitlab.pintechs.com/eh/energy-backend/cap/msg/i18n"
	"golang.org/x/text/language"
	"testing"
	"time"
)

type TestTranslator struct {
	lang string
}

func (t TestTranslator) Translate(lang language.Tag, id i18n.TrID, args ...interface{}) (string, error) {
	if id == "UNKNOWN_ID" {
		if lang == language.AmericanEnglish {
			return fmt.Sprintf("Unknown Error %d, %d, %s", args...), nil
		} else if lang == language.SimplifiedChinese {
			return fmt.Sprintf("未知错误 %d, %d, %s", args...), nil
		}
	}
	return "", i18n.ErrNoTranslateCandidate
}

func TestMsg_GetMessage(t *testing.T) {
	testMsg := Msg{Timestamp: time.Now(), ID: "UNKNOWN_ID", Args: []interface{}{1, 2, "arg3"}, Severity: Critical}
	trEN := "Unknown Error 1, 2, arg3"
	trCN := "未知错误 1, 2, arg3"
	i18n.AddTrCandidates(&TestTranslator{})
	tests := []struct {
		name string
		m    Msg
		lang language.Tag
		want string
	}{
		{"TRANSLATE_CN", testMsg, language.SimplifiedChinese, trCN},
		{"TRANSLATE_EN", testMsg, language.AmericanEnglish, trEN},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.m.GetMessage(tt.lang); got != tt.want || err != nil {
				t.Errorf("Msg.GetMessage() = %v, %v, want %v", got, err, tt.want)
			}
		})
	}
}
