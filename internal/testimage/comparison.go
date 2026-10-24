package testimage

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed testdata/comparison.txt
var comparison string

// Result is one line of comparison.txt.
type Result struct {
	id         string
	components Components
	Hash       string
}

// Comparison is a parsed comparison.txt.
type Comparison struct {
	header  []string
	Results []Result
}

// value returns the rest of the header line starting with key, or "" if
// there is none.
func (c Comparison) value(key string) string {
	for _, h := range c.header {
		if v, ok := strings.CutPrefix(h, key+" "); ok {
			return v
		}
	}

	return ""
}

// Hashes returns every distinct hash in c, in file order. Different cases can
// share a hash, small flat ones most often.
func (c Comparison) Hashes() []string {
	var out []string

	seen := map[string]bool{}
	for _, r := range c.Results {
		if !seen[r.Hash] {
			seen[r.Hash] = true
			out = append(out, r.Hash)
		}
	}

	return out
}

// Comparisons parses comparison.txt: header lines start with "# ", every
// other line is <case id> <x>x<y> <hash>.
func Comparisons() (Comparison, error) {
	var c Comparison

	for n, line := range strings.Split(strings.TrimSuffix(comparison, "\n"), "\n") {
		if h, ok := strings.CutPrefix(line, "# "); ok {
			c.header = append(c.header, h)

			continue
		}

		var r Result

		_, err := fmt.Sscanf(line, "%s %dx%d %s", &r.id, &r.components.X, &r.components.Y, &r.Hash)
		if err != nil {
			return Comparison{}, fmt.Errorf("line %d: %w", n+1, err)
		}

		c.Results = append(c.Results, r)
	}

	return c, nil
}
