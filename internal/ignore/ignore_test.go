package ignore

import "testing"

func TestIsExcludedGlobstarMiddle(t *testing.T) {
	patterns := []string{"**/node_modules/**"}
	cases := map[string]bool{
		"node_modules/foo.js":         true,
		"src/node_modules/foo/bar.js": true,
		"src/foo.js":                  false,
	}
	for path, want := range cases {
		if got := IsExcluded(path, patterns); got != want {
			t.Errorf("IsExcluded(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestIsExcludedSuffixGlob(t *testing.T) {
	patterns := []string{"**/*.min.js"}
	if !IsExcluded("dist/app.min.js", patterns) {
		t.Error("expected dist/app.min.js to be excluded")
	}
	if !IsExcluded("app.min.js", patterns) {
		t.Error("expected app.min.js (root) to be excluded")
	}
	if IsExcluded("app.js", patterns) {
		t.Error("expected app.js to NOT be excluded")
	}
}

func TestIsExcludedBareFilenameMatchesAnyDepth(t *testing.T) {
	patterns := []string{"package-lock.json"}
	if !IsExcluded("package-lock.json", patterns) {
		t.Error("expected root package-lock.json to be excluded")
	}
	if !IsExcluded("packages/app/package-lock.json", patterns) {
		t.Error("expected nested package-lock.json to be excluded via basename match")
	}
}

func TestIsExcludedDefaultExcludesCoverNodeModules(t *testing.T) {
	if !IsExcluded("node_modules/foo/index.js", DefaultExcludes) {
		t.Error("expected default excludes to cover node_modules")
	}
}
