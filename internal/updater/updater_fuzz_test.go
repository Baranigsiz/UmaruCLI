package updater

import (
	"testing"
)

func FuzzCleanVersion(f *testing.F) {
	f.Add("v1.2.3")
	f.Add("1.2.3")
	f.Add("  v2.0.0-beta.1  ")
	f.Add("")
	f.Add("v")
	f.Add("vvv1.0")
	f.Add("random_string_123")

	f.Fuzz(func(t *testing.T, input string) {
		clean := CleanVersion(input)
		// CleanVersion must never panic
		_ = clean
	})
}

func FuzzIsNewerVersion(f *testing.F) {
	f.Add("1.0.0", "1.0.1")
	f.Add("v1.0.0", "v1.1.0")
	f.Add("1.0.0-alpha", "1.0.0-beta")
	f.Add("1.0.0-beta.1", "1.0.0-beta.2")
	f.Add("dev", "v1.0.0")
	f.Add("invalid", "invalid")
	f.Add("", "")
	f.Add("9999999999999999999999999", "1.0.0")
	f.Add("1.0.0...1", "1.0.0.1")

	f.Fuzz(func(t *testing.T, curr, latest string) {
		// Must never panic on arbitrary inputs
		res := IsNewerVersion(curr, latest)

		// If both versions are non-empty and strictly identical, IsNewerVersion should always be false
		if curr != "" && curr == latest {
			if res {
				t.Errorf("IsNewerVersion(%q, %q) returned true for identical versions", curr, latest)
			}
		}
	})
}
