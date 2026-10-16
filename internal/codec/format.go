package codec

import (
	"math"
	"strings"
)

// format quantises the coefficients and writes them as a hash, with each
// channel's DC tie, where it has one, as its DC byte. The quantisation is the
// reference encoder's.
func format(nx, ny int, f [][3]float64, dcTie [3]int) string {
	var b strings.Builder

	// The size digit, the maximum, 4 digits of DC and 2 per AC component.
	b.Grow(4 + 2*nx*ny)
	writeBase83(&b, (nx-1)+(ny-1)*9, 1)

	dc, ac := f[0], f[1:]

	maximum := 1.0

	if len(ac) > 0 {
		actual := 0.0
		for _, c := range ac {
			for _, v := range c {
				actual = max(actual, math.Abs(v))
			}
		}

		q := quantiseMaximum(actual)
		maximum = unquantiseMaximum(q)
		writeBase83(&b, q, 1)
	} else {
		writeBase83(&b, 0, 1)
	}

	var rgb [3]int

	for c := range rgb {
		rgb[c] = dcTie[c]
		if rgb[c] < 0 {
			rgb[c] = toSRGB(dc[c])
		}
	}

	writeBase83(&b, rgb[0]<<16+rgb[1]<<8+rgb[2], 4)

	for _, c := range ac {
		var d [3]int

		for k, v := range c {
			d[k] = quantiseAC(v, maximum)
		}

		writeBase83(&b, d[0]*19*19+d[1]*19+d[2], 2)
	}

	return b.String()
}

// quantiseMaximum returns the digit, 0 to 82, of the largest AC magnitude v,
// in steps of 1/166.
func quantiseMaximum(v float64) int {
	return int(max(0, min(82, math.Floor(float64(v*166)-0.5))))
}

// unquantiseMaximum returns the maximum AC value that digit q stands for.
func unquantiseMaximum(q int) float64 {
	return float64(q+1) / 166
}

// quantiseAC returns the digit, 0 to 18, of AC value v: the square root of
// v/maximum, keeping its sign, in 19 steps from -1 to 1.
func quantiseAC(v, maximum float64) int {
	q := v / maximum
	root := math.Copysign(math.Sqrt(math.Abs(q)), q)

	return int(max(0, min(18, math.Floor(float64(root*9)+9.5))))
}
