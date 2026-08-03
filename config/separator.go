package config

import (
	"fmt"
	"strconv"
	"strings"
)

// parseSeparatorWidth parses the separator's width or width=N argument.
func parseSeparatorWidth(value string) (int, error) {
	value = strings.TrimSpace(value)
	if keyValue := strings.SplitN(value, "=", 2); len(keyValue) == 2 {
		if !strings.EqualFold(keyValue[0], "width") {
			return 0, fmt.Errorf("unknown separator option %q", keyValue[0])
		}
		value = keyValue[1]
	}
	width, err := strconv.Atoi(value)
	if err != nil || width < 1 || width > 512 {
		return 0, fmt.Errorf("separator width must be between 1 and 512")
	}
	return width, nil
}
