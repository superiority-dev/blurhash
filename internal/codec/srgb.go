package codec

import "math"

// cells divides [0, 1) for cellStart; no cell holds more than two
// thresholds.
const cells = 4096

// cellStart[k] is the byte of the linear value k/cells.
var cellStart = func() (s [cells]uint8) {
	q := 0

	for k := range s {
		for q < 255 && thresholds[q] <= float64(k)/cells {
			q++
		}

		s[k] = uint8(q)
	}

	return s
}()

// edges[q] and edges[q+1] are the thresholds below and above byte q, with
// -Inf below 0 and +Inf above 255.
var edges = func() (e [257]float64) {
	e[0], e[256] = math.Inf(-1), math.Inf(1)
	copy(e[1:256], thresholds[:])

	return e
}()

// toSRGB returns the byte of linear value v: the number of thresholds v
// reaches, which rounds the inverse curve half up. It looks up v's cell, then
// steps to the exact byte against the thresholds.
func toSRGB(v float64) int {
	switch {
	case math.IsNaN(v):
		return 255
	case v <= 0:
		return 0
	case v >= 1:
		return 255
	}

	q := int(cellStart[int(v*cells)])

	for v < edges[q] {
		q--
	}

	for v >= edges[q+1] {
		q++
	}

	return q
}
