package i18n

import (
	"sync"

	"golang.org/x/text/language"
)

// Translator interface
type Translator interface {
	Translate(lang language.Tag, id TrID, args ...interface{}) (string, error)
}

type translatorCandidates struct {
	translators []Translator
	trMutex     sync.RWMutex
}

var trCandidates translatorCandidates

func (tc *translatorCandidates) translate(lang language.Tag, id TrID, args ...interface{}) (string, error) {
	tc.trMutex.RLock()
	defer tc.trMutex.RUnlock()
	for _, tr := range tc.translators {
		ret, err := tr.Translate(lang, id, args...)
		if err != nil {
			continue
		}
		return ret, nil

	}
	return "", ErrNoTranslateCandidate
}

func (tc *translatorCandidates) addTranslators(t ...Translator) {
	tc.trMutex.Lock()
	defer tc.trMutex.Unlock()
	tc.translators = append(tc.translators, t...)
}

// AddTrCandidates add message translator candidates
func AddTrCandidates(t ...Translator) {
	trCandidates.addTranslators(t...)
}

// Translate with candidates
func Translate(lang language.Tag, id TrID, args ...interface{}) (string, error) {
	return trCandidates.translate(lang, id, args...)
}
