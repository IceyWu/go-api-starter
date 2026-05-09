package i18n

import "sync"

// Lang represents a language code
type Lang string

const (
	ZhCN Lang = "zh-CN"
	EnUS Lang = "en-US"
)

var (
	defaultLang Lang = ZhCN
	mu          sync.RWMutex
	messages    = map[Lang]map[string]string{}
)

func init() {
	Register(ZhCN, zhCN)
	Register(EnUS, enUS)
}

// Register registers messages for a language
func Register(lang Lang, m map[string]string) {
	mu.Lock()
	defer mu.Unlock()
	messages[lang] = m
}

// SetDefault sets the default language
func SetDefault(lang Lang) {
	mu.Lock()
	defer mu.Unlock()
	defaultLang = lang
}

// T translates an error code to a message in the given language.
// Falls back to default language, then to the code itself.
func T(code string, lang ...Lang) string {
	mu.RLock()
	defer mu.RUnlock()

	l := defaultLang
	if len(lang) > 0 && lang[0] != "" {
		l = lang[0]
	}

	if m, ok := messages[l]; ok {
		if msg, ok := m[code]; ok {
			return msg
		}
	}
	// Fallback to default language
	if l != defaultLang {
		if m, ok := messages[defaultLang]; ok {
			if msg, ok := m[code]; ok {
				return msg
			}
		}
	}
	return code
}
