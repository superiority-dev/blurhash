package codec

import (
	"math"
	"strings"
)

// The component counts a hash can hold on each axis.
const (
	MinComponents = 1
	MaxComponents = 9
)

// A hash is base-83 digits: sizeDigits for the component counts at the
// start, maximumDigits for the largest AC value at maximumAt, dcDigits for
// the average color at dcAt, then acDigits for each AC component from acAt.
const (
	sizeDigits    = 1
	maximumDigits = 1
	dcDigits      = 4
	acDigits      = 2

	maximumAt = sizeDigits
	dcAt      = maximumAt + maximumDigits
	acAt      = dcAt + dcDigits

	// acBase + acDigits*k is where AC component k starts, for k from 1.
	acBase = acAt - acDigits
)

// acLevels is how many values each AC channel quantizes to, from -1 to 1, and
// acMid is the level of 0.
const (
	acLevels = 19
	acMid    = (acLevels - 1) / 2
)

// maximumScale is the steps per unit of the largest AC value: digit q stands
// for (q+1)/maximumScale.
const maximumScale = 166

// maxDC is the largest average color, a byte each of red, green and blue.
const maxDC = 0xffffff

// format quantizes the coefficients and writes them as a hash. The
// quantization is the reference encoder's.
func format(nx, ny int, f [][3]float64) string {
	var b strings.Builder

	b.Grow(acBase + acDigits*nx*ny)
	writeBase83(&b, (nx-1)+(ny-1)*MaxComponents, sizeDigits)

	dc, ac := f[0], f[1:]

	maximum := 1.0

	if len(ac) > 0 {
		actual := 0.0
		for _, c := range ac {
			for _, v := range c {
				actual = max(actual, math.Abs(v))
			}
		}

		q := quantizeMaximum(actual)
		maximum = unquantizeMaximum(q)
		writeBase83(&b, q, maximumDigits)
	} else {
		writeBase83(&b, 0, maximumDigits)
	}

	writeBase83(&b, toSRGB(dc[0])<<16+toSRGB(dc[1])<<8+toSRGB(dc[2]), dcDigits)

	for _, c := range ac {
		var d [3]int

		for k, v := range c {
			d[k] = quantizeAC(v, maximum)
		}

		writeBase83(&b, d[0]*acLevels*acLevels+d[1]*acLevels+d[2], acDigits)
	}

	return b.String()
}

// quantizeMaximum returns the digit, 0 to 82, of the largest AC magnitude v,
// in steps of 1/166.
func quantizeMaximum(v float64) int {
	return int(max(0, min(float64(base-1), math.Floor(float64(v*maximumScale)-0.5))))
}

// unquantizeMaximum returns the maximum AC value that digit q stands for.
func unquantizeMaximum(q int) float64 {
	return float64(q+1) / maximumScale
}

// quantizeAC returns the digit, 0 to 18, of AC value v: the square root of
// v/maximum, keeping its sign, in 19 steps from -1 to 1.
func quantizeAC(v, maximum float64) int {
	q := v / maximum
	root := math.Copysign(math.Sqrt(math.Abs(q)), q)

	return int(max(0, min(acLevels-1, math.Floor(float64(root*acMid)+(acMid+0.5)))))
}
