package ui

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// FormatShortDateTime formats a time in local timezone.
func FormatShortDateTime(t time.Time) string {
	return t.Local().Format(localeTimeFormat())
}

func localeTimeFormat() string {
	lang := extractLang(getLocale())
	datePart := localeDateFormat(lang)

	use24h := viper.GetBool("time-display.use-24h")
	showSec := viper.GetBool("time-display.show-seconds")

	var timePart string
	if use24h {
		timePart = "15:04"
	} else {
		// Check locale default when not explicitly set
		if !viper.IsSet("time-display.use-24h") {
			timePart = localeTimePart(lang)
		} else {
			timePart = "3:04 PM"
		}
	}

	if showSec {
		timePart = addSeconds(timePart)
	}

	return datePart + " " + timePart
}

// addSeconds inserts :05 (seconds) into the time format string.
func addSeconds(fmt string) string {
	// "15:04" → "15:04:05", "3:04 PM" → "3:04:05 PM"
	if i := strings.Index(fmt, " PM"); i >= 0 {
		return fmt[:i] + ":05" + fmt[i:]
	}
	return fmt + ":05"
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
