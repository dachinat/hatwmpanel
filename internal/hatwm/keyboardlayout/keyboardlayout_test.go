package keyboardlayout

import "testing"

func TestCodeUsesConfiguredMappings(t *testing.T) {
	mappings := map[string]string{
		"us": "EN",
		"ge": "KA",
	}
	for layout, want := range map[string]string{
		"us": "EN", " US+RU ": "EN", "ge": "KA",
		"gb": "GB", "custom": "CU", "": "--",
	} {
		if got := Code(layout, mappings); got != want {
			t.Fatalf("Code(%q) = %q, want %q", layout, got, want)
		}
	}
}

func TestCodeNormalizesVariantsAndUsesGenericFallbacks(t *testing.T) {
	for layout, want := range map[string]string{
		" US+RU ":   "US",
		"fr:oss":    "FR",
		"pt(abnt2)": "PT",
		"ქართული":   "ᲥᲐ",
	} {
		if got := Code(layout, nil); got != want {
			t.Fatalf("Code(%q) = %q, want %q", layout, got, want)
		}
	}
}

func TestCodeIgnoresBlankMapping(t *testing.T) {
	if got := Code("us", map[string]string{"us": "  "}); got != "US" {
		t.Fatalf("Code with blank mapping = %q, want US", got)
	}
}
