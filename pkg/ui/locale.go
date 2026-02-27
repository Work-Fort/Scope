package ui

import (
	"os"
	"strings"
	"time"
)

// FormatShortDateTime formats a time in local timezone using the system locale's short date/time style.
func FormatShortDateTime(t time.Time) string {
	return t.Local().Format(localeTimeFormat())
}

func localeTimeFormat() string {
	lang := extractLang(getLocale())

	// 12-hour locales
	switch lang {
	case "en_US", "en_PH":
		return "1/2 3:04 PM"
	case "en_GB", "en_AU", "en_NZ", "en_IE", "en_ZA":
		return "02/01 15:04"
	case "de", "de_DE", "de_AT", "de_CH":
		return "02.01. 15:04"
	case "ja", "ja_JP", "zh", "zh_CN", "zh_TW", "ko", "ko_KR":
		return "01/02 15:04"
	}

	// Region fallbacks
	if strings.HasPrefix(lang, "en") {
		return "1/2 3:04 PM"
	}

	// Most of the world: D/M 24h
	return "02/01 15:04"
}

func getLocale() string {
	for _, env := range []string{"LC_TIME", "LC_ALL", "LANG"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return "en_US.UTF-8"
}

func extractLang(locale string) string {
	if i := strings.Index(locale, "."); i > 0 {
		locale = locale[:i]
	}
	return locale
}
