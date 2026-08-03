package keyboardlayout

import "testing"

func TestCode(t *testing.T) {
	for layout, want := range map[string]string{
		"us": "EN", "gb": "EN", "ge": "GE", "fr": "FR", "ru": "RU",
		"de(nodeadkeys)": "DE", "custom": "CU", "": "--",
	} {
		if got := Code(layout); got != want {
			t.Fatalf("Code(%q) = %q, want %q", layout, got, want)
		}
	}
}

func TestCodeNormalizesVariantsAndUnicodeFallbacks(t *testing.T) {
	for layout, want := range map[string]string{
		" US+RU ":   "EN",
		"fr:oss":    "FR",
		"pt(abnt2)": "PT",
		"ქართული":   "ᲥᲐ",
	} {
		if got := Code(layout); got != want {
			t.Fatalf("Code(%q) = %q, want %q", layout, got, want)
		}
	}
}
