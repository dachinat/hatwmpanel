// Package keyboardlayout formats XKB layout names for panel display.
package keyboardlayout

import "strings"

func Code(layout string, mappings map[string]string) string {
	layout = strings.ToLower(strings.TrimSpace(layout))
	if separator := strings.IndexAny(layout, "(:+"); separator >= 0 {
		layout = layout[:separator]
	}
	if layout == "" {
		return "--"
	}
	if label := strings.TrimSpace(mappings[layout]); label != "" {
		return label
	}
	runes := []rune(strings.ToUpper(layout))
	if len(runes) > 2 {
		runes = runes[:2]
	}
	return string(runes)
}
