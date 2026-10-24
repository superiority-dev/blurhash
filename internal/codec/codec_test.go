package codec

import (
	"image"
	"math"
	"testing"
)

// tolerance is how far coefficients may lie from slowCoefficients: 1,000
// times the largest gap measured, 9.6e-13, on a 2000x1500 random image at
// 9x9.
const tolerance = 1e-9

// checkCoefficients fails t where a coefficient e computes for img lies more
// than tolerance from slowCoefficients'.
func checkCoefficients(t *testing.T, e *Encoder, id string, nx, ny int, img image.Image) {
	t.Helper()

	got := e.coefficients(nx, ny, img)

	for k, want := range slowCoefficients(nx, ny, img) {
		for c := range want {
			if d := math.Abs(got[k][c] - want[c]); d > tolerance {
				t.Fatalf("%s at %dx%d: coefficient %d channel %d is %g from slowCoefficients'",
					id, nx, ny, k, c, d)
			}
		}
	}
}

// slowCoefficients returns the coefficients of img computed the plain way,
// which the Encoder is checked against: each pixel read through At rather
// than linearRow's direct reads, and each component summed straight from the
// formula rather than in two passes. Index j*nx+i is horizontal component i
// and vertical component j.
func slowCoefficients(nx, ny int, img image.Image) [][3]float64 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	pix := make([][3]float64, w*h)

	for y := range h {
		for x := range w {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pix[y*w+x] = [3]float64{linear[r>>8], linear[g>>8], linear[bl>>8]}
		}
	}

	cosX := cosines(nx, w)
	cosY := cosines(ny, h)

	out := make([][3]float64, nx*ny)

	for j := range ny {
		for i := range nx {
			var sum [3]float64

			for y := range h {
				for x := range w {
					basis := cosX[i][x] * cosY[j][y]

					for c, v := range pix[y*w+x] {
						// The product is rounded before it is added, so no target fuses the two.
						sum[c] += float64(basis * v)
					}
				}
			}

			// The average color (i = j = 0) is the plain mean; every other
			// component is doubled, as the cosine transform's normalization
			// requires. The scale is applied once, after the sum.
			normalization := 2.0
			if i == 0 && j == 0 {
				normalization = 1
			}

			scale := normalization / float64(w*h)

			for c := range sum {
				out[j*nx+i][c] = sum[c] * scale
			}
		}
	}

	return out
}

// cosines returns cos(pi*k*p/n) for each component k below count and each
// position p below n.
func cosines(count, n int) [][]float64 {
	out := make([][]float64, count)

	for k := range out {
		out[k] = make([]float64, n)
		for p := range n {
			out[k][p] = cosine(k, p, n)
		}
	}

	return out
}
