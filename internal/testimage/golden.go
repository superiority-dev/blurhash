package testimage

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// Golden checks that lines, one per line, are exactly the file at path in
// fsys; with update it rewrites the file instead.
func Golden(t *testing.T, fsys fs.FS, path string, lines []string, update bool) {
	t.Helper()

	if update {
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		return
	}

	b, err := fs.ReadFile(fsys, path)
	if err != nil {
		t.Fatal(err)
	}

	want := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")

	for i := range min(len(lines), len(want)) {
		if lines[i] != want[i] {
			t.Fatalf("%s line %d: got %q, want %q", path, i+1, lines[i], want[i])
		}
	}

	if len(lines) != len(want) {
		t.Fatalf("%s: got %d lines, want %d", path, len(lines), len(want))
	}
}
