package ui

import (
	"os"
	"strings"
	"time"
)

// TimeFormatOverride allows forcing 12h or 24h time display.
// Set to "12h" or "24h" to override locale detection. Empty uses locale default.
var TimeFormatOverride string

// FormatShortDateTime formats a time in local timezone using the system locale's short date/time style.
func FormatShortDateTime(t time.Time) string {
	return t.Local().Format(localeTimeFormat())
}

func localeTimeFormat() string {
	lang := extractLang(getLocale())
	datePart := localeDateFormat(lang)

	switch TimeFormatOverride {
	case "24h":
		return datePart + " 15:04"
	case "12h":
		return datePart + " 3:04 PM"
	}

	return datePart + " " + localeTimePart(lang)
}

func localeDateFormat(lang string) string {
	switch lang {
	case "de", "de_DE", "de_AT", "de_CH":
		return "02.01."
	case "ja", "ja_JP", "zh", "zh_CN", "zh_TW", "ko", "ko_KR":
		return "01/02"
	case "en_US", "en_PH":
		return "1/2"
	}
	if strings.HasPrefix(lang, "en") {
		return "1/2"
	}
	return "02/01"
}

func localeTimePart(lang string) string {
	switch lang {
	case "en_US", "en_PH":
		return "3:04 PM"
	}
	if strings.HasPrefix(lang, "en") {
		return "3:04 PM"
	}
	return "15:04"
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
