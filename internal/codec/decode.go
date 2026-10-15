package codec

import (
	"fmt"
	"image/color"
	"image/draw"
	"math"
)

// Components returns the numbers of horizontal and vertical components
// hash encodes.
func Components(hash string) (x, y int, err error) {
	if len(hash) < 6 {
		return 0, 0, fmt.Errorf("%w: %d characters, need at least 6", ErrInvalidHash, len(hash))
	}

	size, ok := readBase83(hash[:1])
	if !ok || size > 80 {
		return 0, 0, fmt.Errorf("%w: size %q", ErrInvalidHash, hash[:1])
	}

	x, y = size%9+1, size/9+1
	if n := 4 + 2*x*y; len(hash) != n {
		return 0, 0, fmt.Errorf("%w: %d characters, need %d for %dx%d components",
			ErrInvalidHash, len(hash), n, x, y)
	}

	return x, y, nil
}

// colours writes the linear-light colour of each component of hash into dst,
// index j*nx+i, with the AC amplitudes scaled by punch.
func colours(dst *[81][3]float64, hash string, punch float64) (nx, ny int, err error) {
	nx, ny, err = Components(hash)
	if err != nil {
		return 0, 0, err
	}

	if math.IsNaN(punch) || math.IsInf(punch, 0) {
		return 0, 0, fmt.Errorf("%w: got %v", ErrInvalidPunch, punch)
	}

	q, ok := readBase83(hash[1:2])
	if !ok {
		return 0, 0, fmt.Errorf("%w: maximum %q", ErrInvalidHash, hash[1:2])
	}

	amplitude := unquantiseMaximum(q) * punch

	dc, ok := readBase83(hash[2:6])
	if !ok || dc > 0xffffff {
		return 0, 0, fmt.Errorf("%w: average colour %q", ErrInvalidHash, hash[2:6])
	}

	dst[0] = [3]float64{linear[dc>>16], linear[dc>>8&0xff], linear[dc&0xff]}

	for k := 1; k < nx*ny; k++ {
		ac, ok := readBase83(hash[4+2*k : 6+2*k])
		if !ok || ac >= 19*19*19 {
			return 0, 0, fmt.Errorf("%w: component %d %q", ErrInvalidHash, k, hash[4+2*k:6+2*k])
		}

		dst[k] = [3]float64{
			unquantise(ac/(19*19), amplitude),
			unquantise(ac/19%19, amplitude),
			unquantise(ac%19, amplitude),
		}
	}

	return nx, ny, nil
}

// unquantise returns the AC value of digit q at the given amplitude: the
// square of (q-9)/9, keeping its sign, times the amplitude.
func unquantise(q int, amplitude float64) float64 {
	v := float64(q-9) / 9

	return v * math.Abs(v) * amplitude
}

// Draw decodes hash into dst, filling its bounds, with the components'
// contrast scaled by punch: each pixel summed straight from the formula,
// component by component. It is safe for concurrent use.
func Draw(dst draw.Image, hash string, punch float64) error {
	var c [81][3]float64

	nx, ny, err := colours(&c, hash, punch)
	if err != nil {
		return err
	}

	b := dst.Bounds()
	w, h := b.Dx(), b.Dy()
	cosX := cosines(nx, w)
	cosY := cosines(ny, h)

	for y := range h {
		for x := range w {
			var sum [3]float64

			for j := range ny {
				for i := range nx {
					basis := cosX[i][x] * cosY[j][y]

					for ch, v := range c[j*nx+i] {
						// The product is rounded before it is added, so no target fuses the two.
						sum[ch] += float64(v * basis)
					}
				}
			}

			r, g, bl := toSRGB(sum[0]), toSRGB(sum[1]), toSRGB(sum[2])
			dst.Set(b.Min.X+x, b.Min.Y+y, color.NRGBA{uint8(r), uint8(g), uint8(bl), 255})
		}
	}

	return nil
}
