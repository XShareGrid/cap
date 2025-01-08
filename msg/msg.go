package msg

import (
	"time"

	"github.com/XShareGrid/cap/msg/i18n"
	"golang.org/x/text/language"
)

// Severity Message severity
type Severity int

const (
	// NotSpecified ...
	NotSpecified = iota
	// Normal ..
	Normal
	// Warning ..
	Warning
	// Critical ..
	Critical
)

// Msg msg definition
// member:Time message timestamp
// member:ID message ID
// member:Args message args
// member:Severity message severity
// member:Custom user specific field
type Msg struct {
	Timestamp time.Time
	ID        i18n.TrID
	Args      []interface{}
	Severity  Severity
	Custom    interface{}
}

// GetMessage get translated message string
// lang: language tag
// translator: custom translator, use default candidates if no translator specified
func (m *Msg) GetMessage(lang language.Tag, translators ...i18n.Translator) (string, error) {
	if len(translators) != 0 {
		for _, tr := range translators {
			ret, err := tr.Translate(lang, m.ID, m.Args...)
			if err != nil {
				continue
			}
			return ret, nil
		}
	}
	return i18n.Translate(lang, m.ID, m.Args...)
}
