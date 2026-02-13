package shared_test

import (
	"strings"
	"testing"

	"github.com/riddopic/cc-tools/internal/shared"
)

// assertHasPrefix checks that got starts with the given prefix.
func assertHasPrefix(t *testing.T, got, prefix string) {
	t.Helper()

	if !strings.HasPrefix(got, prefix) {
		t.Errorf("got %v, want prefix %v", got, prefix)
	}
}

// assertHasSuffix checks that got ends with the given suffix.
func assertHasSuffix(t *testing.T, got, suffix string) {
	t.Helper()

	if !strings.HasSuffix(got, suffix) {
		t.Errorf("got %v, want suffix %v", got, suffix)
	}
}

// assertContainsAll checks that got contains every element in wantParts.
func assertContainsAll(t *testing.T, got string, wantParts []string) {
	t.Helper()

	for _, want := range wantParts {
		if !strings.Contains(got, want) {
			t.Errorf("got %v, want to contain %v", got, want)
		}
	}
}

// assertValidHexHash verifies the debug log path ends with an 8-char hex hash
// before the .debug extension.
func assertValidHexHash(t *testing.T, got string) {
	t.Helper()

	parts := strings.Split(got, "-")
	if len(parts) < 3 {
		t.Errorf("got %v, expected at least 3 parts separated by '-'", got)
		return
	}

	lastPart := parts[len(parts)-1]
	hashPart := strings.TrimSuffix(lastPart, ".debug")

	const expectedHashLen = 8
	if len(hashPart) != expectedHashLen {
		t.Errorf("hash part = %v, expected %d hex characters", hashPart, expectedHashLen)
		return
	}

	for _, ch := range hashPart {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			t.Errorf("hash part = %v, contains non-hex character %c", hashPart, ch)
		}
	}
}

func TestGetDebugLogPathForDir(t *testing.T) {
	tests := []struct {
		name         string
		dir          string
		wantContains []string
		wantPrefix   string
		wantSuffix   string
	}{
		{
			name:         "normal directory path",
			dir:          "/home/user/projects/myapp",
			wantContains: []string{"projects-myapp"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "single directory",
			dir:          "/tmp",
			wantContains: []string{"-tmp-"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "deep nested path",
			dir:          "/very/deep/nested/directory/structure/here",
			wantContains: []string{"structure-here"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "path with spaces",
			dir:          "/home/user/my projects/app name",
			wantContains: []string{"my_projects-app_name"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "root directory",
			dir:          "/",
			wantContains: []string{"root"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "empty parts after root",
			dir:          "//",
			wantContains: []string{"root"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
		{
			name:         "trailing slash",
			dir:          "/home/user/project/",
			wantContains: []string{"user-project"},
			wantPrefix:   "/tmp/cc-tools-",
			wantSuffix:   ".debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shared.GetDebugLogPathForDir(tt.dir)

			assertHasPrefix(t, got, tt.wantPrefix)
			assertHasSuffix(t, got, tt.wantSuffix)
			assertContainsAll(t, got, tt.wantContains)
			assertValidHexHash(t, got)
		})
	}
}

func TestGetDebugLogPathForDirConsistency(t *testing.T) {
	// Test that the same directory always produces the same path.
	dir := "/home/user/test/project"

	path1 := shared.GetDebugLogPathForDir(dir)
	path2 := shared.GetDebugLogPathForDir(dir)
	path3 := shared.GetDebugLogPathForDir(dir)

	if path1 != path2 || path2 != path3 {
		t.Errorf("returned different paths for same directory: %v, %v, %v", path1, path2, path3)
	}

	// Test that different directories produce different paths.
	dir2 := "/home/user/test/other"
	path4 := shared.GetDebugLogPathForDir(dir2)

	if path4 == path1 {
		t.Errorf("returned same path for different directories")
	}
}

func TestGetDebugLogPathForDirEdgeCases(t *testing.T) {
	// Test with relative path (should still work).
	path := shared.GetDebugLogPathForDir("relative/path")
	if !strings.HasPrefix(path, "/tmp/cc-tools-") || !strings.HasSuffix(path, ".debug") {
		t.Errorf("with relative path = %v, want /tmp/cc-tools-*.debug", path)
	}

	// Test with path containing multiple slashes.
	path2 := shared.GetDebugLogPathForDir("/home//user///project")
	if !strings.Contains(path2, "user-project") {
		t.Errorf("didn't handle multiple slashes correctly: %v", path2)
	}

	// Test with very long directory name.
	longName := strings.Repeat("a", 100) + "/" + strings.Repeat("b", 100)
	path3 := shared.GetDebugLogPathForDir(longName)
	if !strings.HasPrefix(path3, "/tmp/cc-tools-") || !strings.HasSuffix(path3, ".debug") {
		t.Errorf("with long name = %v, want /tmp/cc-tools-*.debug", path3)
	}
}
