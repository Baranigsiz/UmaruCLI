package generator

import (
	"regexp"
	"strings"
	"testing"
)

var validSlugRegex = regexp.MustCompile(`^[a-z0-9_\-]+$`)

func FuzzSlugify(f *testing.F) {
	// Seed corpus with various inputs
	f.Add("My Awesome Project")
	f.Add("Çalışma Masası 2026")
	f.Add("hello_world-123")
	f.Add("   ---spaced---   ")
	f.Add("!@#$%^&*()_+")
	f.Add("")
	f.Add("İstanbul Şampiyonası ğüşiöç")
	f.Add("../../../etc/passwd")
	f.Add("null\x00byte")
	f.Add("🔥🚀💻")

	f.Fuzz(func(t *testing.T, input string) {
		slug := Slugify(input)

		// 1. Slug must never be empty
		if len(slug) == 0 {
			t.Errorf("Slugify(%q) produced empty string", input)
		}

		// 2. Slug must only contain allowed characters [a-z0-9_\-]
		if !validSlugRegex.MatchString(slug) {
			t.Errorf("Slugify(%q) = %q contains invalid characters", input, slug)
		}

		// 3. Slug must not start or end with a hyphen
		if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
			t.Errorf("Slugify(%q) = %q has leading or trailing hyphen", input, slug)
		}

		// 4. Slugify must be idempotent: Slugify(Slugify(x)) == Slugify(x)
		secondSlug := Slugify(slug)
		if secondSlug != slug {
			t.Errorf("Slugify is not idempotent: Slugify(%q) = %q, then Slugify(%q) = %q", input, slug, slug, secondSlug)
		}
	})
}
