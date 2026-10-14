package codec

import (
	"image"
	"math"
)

// Encode returns the BlurHash of img at nx by ny components, each between 1
// and 9, for an image with pixels. It is safe for concurrent use.
func Encode(nx, ny int, img image.Image) string {
	return format(nx, ny, factors(nx, ny, img))
}

// factors returns the cosine transform coefficients of img in linear light,
// index j*nx+i for horizontal component i and vertical component j.
func factors(nx, ny int, img image.Image) [][3]float64 {
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

			// The average colour (i = j = 0) is the plain mean; every other
			// component is doubled, as the cosine transform's normalisation
			// requires. The scale is applied once, after the sum.
			normalisation := 2.0
			if i == 0 && j == 0 {
				normalisation = 1
			}

			scale := normalisation / float64(w*h)

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
			out[k][p] = math.Cos(math.Pi * float64(k) * float64(p) / float64(n))
		}
	}

	return out
}
