// Package keyboardlayout formats XKB layout names for panel display.
package keyboardlayout

import "strings"

var languageCodes = map[string]string{
	"us": "EN", "gb": "EN",
	"ge": "GE",
	"fr": "FR",
	"ru": "RU",
	"de": "DE",
	"es": "ES",
	"it": "IT",
	"pt": "PT",
	"ua": "UA",
	"tr": "TR",
	"gr": "GR",
	"pl": "PL",
	"cz": "CZ",
	"se": "SE",
	"no": "NO",
	"dk": "DK",
	"fi": "FI",
	"nl": "NL",
	"be": "BE",
	"ch": "CH",
	"jp": "JA",
	"kr": "KO",
	"cn": "ZH",
}

func Code(layout string) string {
	layout = strings.ToLower(strings.TrimSpace(layout))
	if separator := strings.IndexAny(layout, "(:+"); separator >= 0 {
		layout = layout[:separator]
	}
	if code := languageCodes[layout]; code != "" {
		return code
	}
	if layout == "" {
		return "--"
	}
	runes := []rune(strings.ToUpper(layout))
	if len(runes) > 2 {
		runes = runes[:2]
	}
	return string(runes)
}
