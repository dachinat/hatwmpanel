package native

import (
	"strings"
	"time"
)

// Format renders a time using the strftime-like tokens supported by HatWMPanel.
func formatClock(at time.Time, format string) string {
	if format == "" {
		format = "%a %d %b  %H:%M"
	}
	replacements := map[string]string{
		"%a": "Mon", "%A": "Monday", "%b": "Jan", "%B": "January",
		"%d": "02", "%m": "01", "%Y": "2006", "%y": "06",
		"%H": "15", "%I": "03", "%M": "04", "%S": "05",
		"%p": "PM", "%z": "-0700",
	}
	const percentMarker = "\u0001PERCENT\u0001"
	format = strings.ReplaceAll(format, "%%", percentMarker)
	for token, layout := range replacements {
		format = strings.ReplaceAll(format, token, at.Format(layout))
	}
	return strings.ReplaceAll(format, percentMarker, "%")
}
