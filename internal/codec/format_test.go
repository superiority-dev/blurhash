package codec

import (
	"math"
	"math/rand/v2"
	"testing"
)

// stable returns quantise(v) and whether quantise gives the same digit
// everywhere within e of v, checking both ends of the interval widened by
// one ulp each. The tests hold format's cheaper checks to it.
func stable(quantise func(float64) int, v, e float64) (int, bool) {
	d := quantise(v)
	if e == 0 {
		return d, true
	}

	lo := quantise(math.Nextafter(v-e, math.Inf(-1)))
	hi := quantise(math.Nextafter(v+e, math.Inf(1)))

	return d, lo == hi
}

// TestCheapStable holds format's check, quantising v-e and v+e as computed,
// to stable: wherever the check accepts e, stable accepts e/2, the bound
// format's e is twice.
func TestCheapStable(t *testing.T) {
	quantisers := []struct {
		name string
		q    func(float64) int
	}{
		{"maximum", quantiseMaximum},
		{"DC", toSRGB},
		{"AC", func(v float64) int { return quantiseAC(v, 0.3) }},
	}

	r := rand.New(rand.NewPCG(5, 0))

	for _, c := range quantisers {
		name, q := c.name, c.q

		for range 200000 {
			v := (2*r.Float64() - 1) * 0.6
			e := []float64{1e-15, 1e-12, 1e-9, 1e-5}[r.IntN(4)]

			if r.IntN(2) == 0 {
				// Land v near a step of the quantiser.
				d := q(v)
				for step := e; step < 1 && q(v+step) == d; step *= 2 {
					v += step / 2
				}
			}

			cheap := q(v-e) == q(v+e)
			_, sure := stable(q, v, e/2)

			if cheap && !sure {
				t.Fatalf("%s: v %v, e %v: check accepts, stable at e/2 refuses", name, v, e)
			}
		}
	}
}

// TestFormatGuard names format's three guard checks, on the maximum, the DC
// and the AC digits. Each must reject a value on a step of its quantiser and
// accept the same digit off the step, or a DC on a step that is an exact
// tie, with the hash format gives unguarded.
func TestFormatGuard(t *testing.T) {
	const e = 1e-12

	maximum := unquantiseMaximum(quantiseMaximum(0.3))

	var (
		dcStep  = thresholds[99] // toSRGB steps from 99 to 100
		maxStep = 40.5 / 166     // quantiseMaximum steps from 39 to 40
		acStep  = maximum / 4    // quantiseAC steps from 13 to 14
		dc      = [3]float64{0.25, 0.25, 0.25}
		none    = [3]int{-1, -1, -1}
	)

	for _, c := range []struct {
		name  string
		f     [][3]float64
		dcTie [3]int
		ok    bool
	}{
		{"maximum on a step", [][3]float64{dc, {maxStep, 0, 0}}, none, false},
		{"maximum off a step", [][3]float64{dc, {40.0 / 166, 0, 0}}, none, true},
		{"DC on a step", [][3]float64{{dcStep, 0.25, 0.25}}, none, false},
		{"DC on a step, an exact tie", [][3]float64{{dcStep, 0.25, 0.25}}, [3]int{100, -1, -1}, true},
		{"DC off a step", [][3]float64{{(thresholds[99] + thresholds[100]) / 2, 0.25, 0.25}}, none, true},
		{"AC on a step", [][3]float64{dc, {0.3, acStep, 0}}, none, false},
		{"AC off a step", [][3]float64{dc, {0.3, maximum * 0.3, 0}}, none, true},
	} {
		nx := len(c.f)

		hash, ok := format(nx, 1, c.f, [3]float64{e, e, e}, c.dcTie)
		if ok != c.ok {
			t.Errorf("%s: format reports %v, want %v", c.name, ok, c.ok)

			continue
		}

		if want, _ := format(nx, 1, c.f, [3]float64{}, c.dcTie); ok && hash != want {
			t.Errorf("%s: format gives %s, unguarded %s", c.name, hash, want)
		}
	}
}
