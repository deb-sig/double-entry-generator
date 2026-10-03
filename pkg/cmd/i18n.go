package cmd

import (
	"os"
	"strings"
)

func msg(en, zh string) string {
	if isChineseLocale() {
		return zh
	}
	return en
}

func isChineseLocale() bool {
	for _, name := range []string{"DEG_LANG", "LC_ALL", "LC_MESSAGES", "LANG"} {
		value := strings.ToLower(os.Getenv(name))
		if strings.HasPrefix(value, "zh") || strings.Contains(value, ".zh") {
			return true
		}
	}
	return false
}
