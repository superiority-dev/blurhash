package codec

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"sync"
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

// readColors writes the linear-light color of each component of hash into dst,
// index j*nx+i, with the AC amplitudes scaled by punch.
func readColors(dst *[81][3]float64, hash string, punch float64) (nx, ny int, err error) {
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

	amplitude := unquantizeMaximum(q) * punch

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
			unquantize(ac/(19*19), amplitude),
			unquantize(ac/19%19, amplitude),
			unquantize(ac%19, amplitude),
		}
	}

	return nx, ny, nil
}

// unquantize returns the AC value of digit q at the given amplitude: the
// square of (q-9)/9, keeping its sign, times the amplitude.
func unquantize(q int, amplitude float64) float64 {
	v := float64(q-9) / 9

	return v * math.Abs(v) * amplitude
}

// decoders holds the Decoders behind Draw.
var decoders = sync.Pool{New: func() any { return new(Decoder) }}

// Draw decodes hash into dst, filling its bounds, with the components'
// contrast scaled by punch, using a Decoder from a pool. It is safe for
// concurrent use.
func Draw(dst draw.Image, hash string, punch float64) error {
	d := decoders.Get().(*Decoder)
	defer decoders.Put(d)

	return d.Draw(dst, hash, punch)
}

// A Decoder decodes BlurHashes and keeps its cosine tables between calls. The
// zero value is ready to use. A Decoder is not safe for concurrent use.
type Decoder struct {
	colors [81][3]float64
	rows   [2][9][3]float64
	tables cache
}

// Draw decodes hash into dst, filling its bounds, with the components'
// contrast scaled by punch.
func (d *Decoder) Draw(dst draw.Image, hash string, punch float64) error {
	nx, ny, err := readColors(&d.colors, hash, punch)
	if err != nil {
		return err
	}

	b := dst.Bounds()
	if b.Empty() {
		return nil
	}

	w, h := b.Dx(), b.Dy()

	p := painter{dst: dst, b: b}

	switch dst := dst.(type) {
	case *image.NRGBA:
		p.pix, p.stride = dst.Pix[dst.PixOffset(b.Min.X, b.Min.Y):], dst.Stride
	case *image.RGBA:
		p.pix, p.stride = dst.Pix[dst.PixOffset(b.Min.X, b.Min.Y):], dst.Stride
	}

	cosX := d.tables.get(w, nx)
	cosY := d.tables.get(h, ny)

	// Rows y and h-y share the products of each vertical component; odd
	// components change sign between them.
	for y := 0; 2*y <= h; y++ {
		d.sumRows(nx, ny, cosY, y)
		p.drawRow(cosX, d.rows[0][:nx], y)

		if y > 0 && 2*y != h {
			p.drawRow(cosX, d.rows[1][:nx], h-y)
		}
	}

	return nil
}

// sumRows sets d.rows[0] to the horizontal component values of row y and
// d.rows[1] to those of row h-y.
func (d *Decoder) sumRows(nx, ny int, cosY [][]float64, y int) {
	for i := range nx {
		var er, eg, eb, or, og, ob float64

		for j := 0; j < ny; j += 2 {
			cy, v := cosY[j][y], d.colors[j*nx+i]
			// Each product is rounded before it is added, so no target fuses the two.
			er += float64(cy * v[0])
			eg += float64(cy * v[1])
			eb += float64(cy * v[2])
		}

		for j := 1; j < ny; j += 2 {
			cy, v := cosY[j][y], d.colors[j*nx+i]
			or += float64(cy * v[0])
			og += float64(cy * v[1])
			ob += float64(cy * v[2])
		}

		d.rows[0][i] = [3]float64{er + or, eg + og, eb + ob}
		d.rows[1][i] = [3]float64{er - or, eg - og, eb - ob}
	}
}
