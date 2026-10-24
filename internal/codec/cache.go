package codec

import (
	"math"
	"slices"

	"github.com/superiority-dev/blurhash/internal/fdlibm"
)

// cache keeps the cosine tables of the dimensions encoded most recently, so
// a stream of images in a few sizes builds each table once.
type cache struct {
	tables []table // most recent first
	floats int
}

// table holds cos(pi*k*p/n) at cos[k][p] for p up to n/2; the mirror
// symmetry gives the rest.
type table struct {
	n   int
	cos [][]float64
}

const (
	cacheSlots = 8 // distinct side lengths

	// cacheFloats is 64 KiB, which holds every thumbnail-scale table. A
	// larger table, such as 9 rows at 1920, is rebuilt on each call.
	cacheFloats = 8192
)

// get returns cos(pi*k*p/side) for each k below components and p up to side/2.
func (c *cache) get(side, components int) [][]float64 {
	at := slices.IndexFunc(c.tables, func(t table) bool { return t.n == side })

	var t table
	if at >= 0 {
		t = c.tables[at]
		c.tables = slices.Delete(c.tables, at, at+1)
		c.floats -= t.floats()
	} else {
		t.n = side
	}

	for k := len(t.cos); k < components; k++ {
		row := make([]float64, side/2+1)
		for p := range row {
			row[p] = cosine(k, p, side)
		}

		t.cos = append(t.cos, row)
	}

	// A table over the whole budget is built on each call, not cached.
	if t.floats() > cacheFloats {
		return t.cos
	}

	c.tables = slices.Insert(c.tables, 0, t)
	c.floats += t.floats()

	for len(c.tables) > cacheSlots || c.floats > cacheFloats {
		last := len(c.tables) - 1
		c.floats -= c.tables[last].floats()
		c.tables = slices.Delete(c.tables, last, last+1)
	}

	return t.cos
}

// floats returns the number of float64s t holds.
func (t table) floats() int {
	return len(t.cos) * (t.n/2 + 1)
}

// cosine returns cos(pi*k*p/n), the basis value every table holds.
func cosine(k, p, n int) float64 {
	return fdlibm.Cos(math.Pi * float64(k) * float64(p) / float64(n))
}
