package codec

import (
	"math"
	"testing"
)

// TestCache checks get against cosine over a stream of sizes, and the
// cache's bookkeeping after every call.
func TestCache(t *testing.T) {
	var c cache

	// Tiny and odd sizes; 32 again while it is cached; more sizes than
	// cacheSlots, so tables are evicted, 24 among them before it returns; and
	// 2000, whose 9 rows are over cacheFloats.
	for _, n := range []int{1, 2, 3, 32, 24, 32, 21, 18, 100, 7, 9, 11, 13, 32, 2000, 24} {
		for _, count := range []int{1, 4, 9} {
			var cached [][]float64
			for _, tb := range c.tables {
				if tb.n == n {
					cached = tb.cos
				}
			}

			cos := c.get(n, count)

			if cached != nil && &cos[0][0] != &cached[0][0] {
				t.Fatalf("n %d: the cached table was rebuilt", n)
			}

			for k := range count {
				for p := range n/2 + 1 {
					got, want := cos[k][p], cosine(k, p, n)
					if math.Float64bits(got) != math.Float64bits(want) {
						t.Fatalf("n %d: table[%d][%d] = %v, cosine gives %v", n, k, p, got, want)
					}
				}
			}

			if len(c.tables) > cacheSlots || c.floats > cacheFloats {
				t.Fatalf("n %d: %d tables, %d floats", n, len(c.tables), c.floats)
			}

			floats := 0
			for _, tb := range c.tables {
				floats += tb.floats()
			}

			if floats != c.floats {
				t.Fatalf("n %d: counted %d floats, cache says %d", n, floats, c.floats)
			}
		}
	}

	if c.tables[0].n != 24 {
		t.Errorf("most recent table has n %d, want 24", c.tables[0].n)
	}

	for _, tb := range c.tables {
		if tb.n == 2000 {
			t.Errorf("a table of %d floats is cached, over the budget of %d", tb.floats(), cacheFloats)
		}
	}
}
