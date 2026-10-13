package codec

import (
	"math"
	"math/big"
	"math/rand/v2"
	"sort"
	"testing"
)

const exactPrec = 512

// power12over5 returns (num/den)^(12/5), computed to exactPrec bits and
// rounded to the nearest float64.
func power12over5(num, den int64) float64 {
	n := new(big.Int).Exp(big.NewInt(num), big.NewInt(12), nil)
	d := new(big.Int).Exp(big.NewInt(den), big.NewInt(12), nil)
	x := new(big.Float).SetPrec(exactPrec).SetRat(new(big.Rat).SetFrac(n, d))
	f, _ := x.Float64()
	y := new(big.Float).SetPrec(exactPrec).SetFloat64(math.Pow(f, 0.2))

	// Newton's method for the fifth root of x, from math.Pow's estimate. Each
	// step about doubles the correct bits, so 20 is far more than exactPrec needs.
	for range 20 {
		y4 := new(big.Float).SetPrec(exactPrec).Mul(y, y)
		y4.Mul(y4, y4)
		t := new(big.Float).SetPrec(exactPrec).Quo(x, y4)
		y.Mul(y, big.NewFloat(4)).Add(y, t).Quo(y, big.NewFloat(5))
	}

	r, _ := y.Float64()

	return r
}

func rational(num, den int64) float64 {
	f, _ := big.NewRat(num, den).Float64()
	return f
}

// TestLinear checks each entry is the sRGB-to-linear value of its byte,
// correctly rounded: v/3294.6 = 5v/16473 up to byte 10, where v/255 <= 0.04045,
// and ((v/255+0.055)/1.055)^2.4 = ((40v+561)/10761)^(12/5) above.
func TestLinear(t *testing.T) {
	for v := range int64(256) {
		want := rational(5*v, 16473)
		if v > 10 {
			want = power12over5(40*v+561, 10761)
		}

		if linear[v] != want {
			t.Errorf("linear[%d] = %x, exact value rounds to %x", v, linear[v], want)
		}
	}
}

// TestThresholds checks each entry is the linear value at which the inverse
// curve reaches k-0.5, correctly rounded: (k-0.5)/3294.6 = 5(2k-1)/32946 up to
// k = 10, on the linear branch, and ((40k+541)/10761)^(12/5) above.
func TestThresholds(t *testing.T) {
	for k := int64(1); k <= 255; k++ {
		want := rational(5*(2*k-1), 32946)
		if k > 10 {
			want = power12over5(40*k+541, 10761)
		}

		if thresholds[k-1] != want {
			t.Errorf("thresholds[%d] = %x, exact value rounds to %x", k-1, thresholds[k-1], want)
		}
	}
}

// TestLinearFormula bounds the gap between the table and the float64 formula
// it replaces, whose own rounding before the power reaches 7 ulp.
func TestLinearFormula(t *testing.T) {
	for v := range 256 {
		f := float64(v) / 255

		formula := f / 12.92
		if f > 0.04045 {
			formula = math.Pow((f+0.055)/1.055, 2.4)
		}

		ulp := math.Abs(formula-linear[v]) / (math.Nextafter(linear[v], 2) - linear[v])
		if ulp > 8 {
			t.Errorf("linear[%d] is %.0f ulp from the formula", v, ulp)
		}
	}
}

func TestToSRGB(t *testing.T) {
	for k := 1; k <= 255; k++ {
		if got := toSRGB(thresholds[k-1]); got != k {
			t.Errorf("toSRGB(thresholds[%d]) = %d, want %d", k-1, got, k)
		}

		if got := toSRGB(math.Nextafter(thresholds[k-1], 0)); got != k-1 {
			t.Errorf("toSRGB just below thresholds[%d] = %d, want %d", k-1, got, k-1)
		}
	}

	// The two ends, and values past them.
	for _, c := range []struct {
		v    float64
		want int
	}{{-1, 0}, {0, 0}, {1, 255}, {2, 255}} {
		if got := toSRGB(c.v); got != c.want {
			t.Errorf("toSRGB(%v) = %d, want %d", c.v, got, c.want)
		}
	}
}

// searchSRGB returns the number of thresholds v reaches, by binary search.
func searchSRGB(v float64) int {
	return sort.Search(len(thresholds), func(k int) bool { return thresholds[k] > v })
}

// TestToSRGBSearch checks toSRGB against searchSRGB at every threshold, one
// ulp either side of it, every cell edge and random values.
func TestToSRGBSearch(t *testing.T) {
	var vs []float64

	for _, th := range thresholds {
		vs = append(vs, th, math.Nextafter(th, 0), math.Nextafter(th, 2))
	}

	for k := range cells + 1 {
		v := float64(k) / cells
		vs = append(vs, v, math.Nextafter(v, -1), math.Nextafter(v, 2))
	}

	// Random values from -0.1 to 1.1, past both ends.
	r := rand.New(rand.NewPCG(3, 0))
	for range 100000 {
		vs = append(vs, r.Float64()*1.2-0.1)
	}

	vs = append(vs,
		math.NaN(), math.Inf(1), math.Inf(-1),
		0, math.Copysign(0, -1), 1, -1e-300, 1e-300,
	)

	for _, v := range vs {
		if got, want := toSRGB(v), searchSRGB(v); got != want {
			t.Fatalf("toSRGB(%v) = %d, searchSRGB gives %d", v, got, want)
		}
	}
}

// TestCells checks no cell holds more than two thresholds.
func TestCells(t *testing.T) {
	for k := range cells {
		lo, hi := float64(k)/cells, float64(k+1)/cells
		n := 0

		for _, th := range thresholds {
			if th >= lo && th < hi {
				n++
			}
		}

		if n > 2 {
			t.Fatalf("cell %d holds %d thresholds", k, n)
		}
	}
}
