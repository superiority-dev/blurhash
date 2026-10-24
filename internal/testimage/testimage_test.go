package testimage

import "testing"

// TestComparisonMatchesCases requires comparison.txt to be made from today's
// cases: the same fingerprint, count, and IDs and components in order.
func TestComparisonMatchesCases(t *testing.T) {
	file, err := Comparisons()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := fingerprint(), file.value("fingerprint"); got != want {
		t.Fatalf("images have fingerprint %s, comparison.txt was made from %s, so it is out of date", got, want)
	}

	n := 0

	for _, c := range Cases() {
		for _, comp := range c.Components {
			if n >= len(file.Results) {
				t.Fatalf("comparison.txt ends at %d results", n)
			}

			if r := file.Results[n]; r.id != c.ID || r.components != comp {
				t.Fatalf("result %d is %s %v, want %s %v", n, r.id, r.components, c.ID, comp)
			}

			n++
		}
	}

	if n != len(file.Results) {
		t.Fatalf("comparison.txt has %d results, cases have %d", len(file.Results), n)
	}
}
