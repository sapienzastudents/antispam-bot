package i18n

import (
	"embed"
	"errors"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// l10n contains all JSON files under data directory. Each file is one language
// translations.
//go:embed data/*
var l10n embed.FS

// Bundle stores a set of messages. Can be initialized with New.
type Bundle struct {
	logger logrus.FieldLogger

	// Parsed JSON files from l10n variable. The first key is language code
	// extracted from JSON's file name, second key is the source string to be
	// translated.
	strings map[string]map[string]string
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
	b.strings = make(map[string]map[string]string)

	// Populates b.strings reading from l10n.
	files, err := l10n.ReadDir("data")
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON's directory: %w", err)
	}
	for _, entry := range files {
		// data/ should contains only JSON files!
		if entry.IsDir() {
			return nil, errors.New("unexpected directory under JSON's i18n files")
		}

		// Extract language code.
		lang := entry.Name()
		if len(lang) < 2 {
			return nil, fmt.Errorf("unexpected non-JSON file under data i18n directory: %q", lang)
		}
		lang = lang[:2]

		// Read and parse JSON file.
		path := "data/"+entry.Name()
		data, err := l10n.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON file: %w", err)
		}
		var x map[string]string
		if err := json.Unmarshal(data, &x); err != nil {
			return nil, fmt.Errorf("failed to decode %q strings: %w", path, err)
		}
		b.strings[lang] = x
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

	if _, ok := b.strings[lang]; !ok {
		logger.Warn("Language not supported")
		return text
	}

	if translation, ok := b.strings[lang][text]; ok {
		return translation
	}
	logger.Warn("Missing translated string")
	return text
}
