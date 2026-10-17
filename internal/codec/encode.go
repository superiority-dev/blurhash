package codec

import (
	"image"
	"math"

	"github.com/superiority-dev/blurhash/internal/fdlibm"
)

// Encode returns the BlurHash of img at nx by ny components, each between 1
// and 9, for an image with pixels. It is safe for concurrent use.
func Encode(nx, ny int, img image.Image) string {
	p := read(img)

	return format(nx, ny, factors(nx, ny, p), p.dcTie)
}

// pixels is an image in linear light, row by row, with the DC byte of each
// channel whose DC is an exact tie, or -1.
type pixels struct {
	w, h   int
	linear [][3]float64
	dcTie  [3]int
}

// read converts img to linear light and finds each channel's DC tie.
func read(img image.Image) pixels {
	b := img.Bounds()
	p := pixels{w: b.Dx(), h: b.Dy(), linear: make([][3]float64, b.Dx()*b.Dy())}

	var low, white [3]uint64

	rational := [3]bool{true, true, true}

	for y := range p.h {
		for x := range p.w {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()

			for c, v := range [3]uint32{r >> 8, g >> 8, bl >> 8} {
				p.linear[y*p.w+x][c] = linear[v]

				switch {
				case v <= 10:
					low[c] += uint64(v)
				case v == 255:
					white[c]++
				default:
					rational[c] = false
				}
			}
		}
	}

	for c := range p.dcTie {
		p.dcTie[c] = -1
		if rational[c] {
			p.dcTie[c] = dcTie(low[c], white[c], uint64(p.w*p.h))
		}
	}

	return p
}

// dcTie returns the DC byte of a channel whose bytes are all at most 10 or
// equal to 255, or -1 if that DC is not an exact tie. Those bytes have the
// rational linear values 5v/16473 and 1, so with s the sum of the low bytes,
// w the count of white ones and n the pixel count, the inverse curve's value
// plus one half is (10s + 32946w + 5n) / 10n on its linear branch, which
// holds up to 10.81.
func dcTie(s, w, n uint64) int {
	num, den := 10*s+32946*w+5*n, 10*n
	if num%den != 0 || num/den > 10 {
		return -1
	}

	return int(num / den)
}

// factors returns the cosine transform coefficients of p, index j*nx+i for
// horizontal component i and vertical component j.
func factors(nx, ny int, p pixels) [][3]float64 {
	w, h := p.w, p.h

	cosX := cosines(nx, w)
	cosY := cosines(ny, h)

	out := make([][3]float64, nx*ny)

	for j := range ny {
		for i := range nx {
			var sum [3]float64

			for y := range h {
				for x := range w {
					basis := cosX[i][x] * cosY[j][y]

					for c, v := range p.linear[y*w+x] {
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
			out[k][p] = cosine(k, p, n)
		}
	}

	return out
}

// cosine returns cos(pi*k*p/n), the basis value every table holds.
func cosine(k, p, n int) float64 {
	return fdlibm.Cos(math.Pi * float64(k) * float64(p) / float64(n))
}
