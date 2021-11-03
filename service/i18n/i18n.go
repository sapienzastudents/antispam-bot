package i18n

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

//go:embed data/it.json
var itRawJSON []byte

// Bundle stores a set of messages. Can be initialized with New.
type Bundle struct {
	logger logrus.FieldLogger

	itStrings map[string]string
}

// New returns a new Bundle instance that can be used to get translated strings.
//
// It returns an error if logger is nil or if some translation file "xx.json"
// under "data/" package's subfolder cannot be parsed.
func New(logger logrus.FieldLogger) (*Bundle, error) {
	if logger == nil {
		return nil, errors.New("logger is required")
	}

	b := Bundle{logger: logger}
	if err := json.Unmarshal(itRawJSON, &b.itStrings); err != nil {
		return nil, fmt.Errorf("failed to decode it.json strings: %w", err)
	}
	return &b, nil
}

// Returns the translation for the given text to corresponding lang in IETF
// two-letters language codes.
//

// Currently supported languages codes are "it" and "en". On other languages
// code or missing translated string it returns the source string and prints a
// warning on logger.
func (b *Bundle) T(lang, text string) string {
	if lang == "en" {
		return text
	}

	logger := b.logger.WithFields(logrus.Fields{"text": text, "lang": lang})

	if lang == "it" {
		if translation, ok := b.itStrings[text]; ok {
			return translation
		}
		logger.Warn("Failed to find translation for given string")
		return text
	}

	logger.Warn("Language not supported")
	return text
}
