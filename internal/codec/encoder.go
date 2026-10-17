package codec

import (
	"image"
	"sync"
)

// encoders holds the Encoders behind Encode.
var encoders = sync.Pool{New: func() any { return new(Encoder) }}

// Encode returns the BlurHash of img at nx by ny components, each between 1
// and 9, for an image with pixels. It is safe for concurrent use.
func Encode(nx, ny int, img image.Image) string {
	e := encoders.Get().(*Encoder)
	hash := e.Encode(nx, ny, img)
	e.trim()
	encoders.Put(e)

	return hash
}

// An Encoder computes BlurHashes and keeps its buffers and cosine tables
// between calls. The zero value is ready to use. An Encoder is not safe for
// concurrent use.
type Encoder struct {
	line, sum, dif, rows, out [][3]float64
	result                    [9][3]float64 // mirror's output, for up to 9 components
	tables                    cache
}

// Encode returns the BlurHash of img at nx by ny components, each between 1
// and 9, for an image with pixels. Where the guard in format cannot vouch
// for a digit, it takes the coefficients from factors instead.
func (e *Encoder) Encode(nx, ny int, img image.Image) string {
	f, bound, dcTie := e.coefficients(nx, ny, img)
	if hash, ok := format(nx, ny, f, bound, dcTie); ok {
		return hash
	}

	p := read(img)
	hash, _ := format(nx, ny, factors(nx, ny, p), [3]float64{}, p.dcTie)

	return hash
}

// trim drops the buffers once they hold over 64 KiB, at 24 bytes a
// [3]float64, so a pooled Encoder does not keep what one large image needed.
func (e *Encoder) trim() {
	if 24*(cap(e.line)+cap(e.sum)+cap(e.dif)+cap(e.rows)+cap(e.out)) > 64<<10 {
		e.line, e.sum, e.dif, e.rows, e.out = nil, nil, nil, nil, nil
	}
}

// coefficients returns the coefficients factors returns, computed by a pass
// along each row and then one down each column, with each channel's DC tie
// and a bound on how far its coefficients can lie from factors'.
//
// Both passes use the mirror symmetry of the basis,
// cos(pi*i*(w-x)/w) = (-1)^i * cos(pi*i*x/w), so positions x and w-x enter
// even components as a sum and odd ones as a difference.
//
// The bound: with u = 2^-53, N = w*h and A the channel's sum of linear
// values, factors is within (N+1)u*A + 2e*A of the exact sum, where e = 103u
// bounds each cosine's error (4u relative on an argument below 8*pi, plus
// one ulp from fdlibm.Cos). The two passes are within (w+h+12)u*A + 2e*A,
// an upper bound chosen for simplicity and loose by about a factor of two.
// The common scale adds 2u*A. The bound is twice the total, with 512 for
// 4e/u and the constant terms. Its size rests on this derivation, not on a
// test.
func (e *Encoder) coefficients(nx, ny int, img image.Image) ([][3]float64, [3]float64, [3]int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	cosX := e.tables.get(w, nx)
	cosY := e.tables.get(h, ny)

	e.line = grow(e.line, w)
	e.sum = grow(e.sum, max(w, h)/2)
	e.dif = grow(e.dif, max(w, h)/2)
	e.rows = grow(e.rows, nx*h)
	e.out = grow(e.out, nx*ny)

	var t ties

	for y := range h {
		t.read(img, y, e.line)

		for i, v := range e.mirror(e.line, cosX[:nx]) {
			e.rows[i*h+y] = v
		}
	}

	// The DC is the plain mean; every AC component is doubled.
	n := float64(w * h)
	dcScale, acScale := 1/n, 2/n

	var total [3]float64 // the DC before scaling: each channel's sum

	for i := range nx {
		for j, v := range e.mirror(e.rows[i*h:(i+1)*h], cosY[:ny]) {
			scale := acScale
			if i == 0 && j == 0 {
				scale = dcScale
				total = v
			}

			e.out[j*nx+i] = [3]float64{v[0] * scale, v[1] * scale, v[2] * scale}
		}
	}

	var bound [3]float64

	// 2^-51 is 4u: twice the bound, times the AC scale's 2.
	k := float64(w*h+w+h+512) / float64(w*h) * 0x1p-51
	for c := range bound {
		bound[c] = k * total[c]
	}

	return e.out, bound, t.dc(w * h)
}

// mirror returns, for each k below len(cos), the sum over x of v[x] times
// cos(pi*k*x/n), where n is len(v) and cos[k] holds positions up to n/2.
func (e *Encoder) mirror(v [][3]float64, cos [][]float64) [][3]float64 {
	n := len(v)
	m := (n - 1) / 2 // the pairs x and n-x, for x from 1 to m
	sum, dif := e.sum[:m], e.dif[:m]

	// Position 0 has no mirror image, and its cosine is 1 at every k.
	fr, fg, fb := v[0][0], v[0][1], v[0][2]

	// The k = 0 result is the plain sum, since every cosine is 1 there.
	dr, dg, db := fr, fg, fb

	for k := range m {
		lo, hi := v[k+1], v[n-1-k] // a position and its mirror image
		sr, sg, sb := lo[0]+hi[0], lo[1]+hi[1], lo[2]+hi[2]
		sum[k] = [3]float64{sr, sg, sb}
		dif[k] = [3]float64{lo[0] - hi[0], lo[1] - hi[1], lo[2] - hi[2]}
		dr, dg, db = dr+sr, dg+sg, db+sb
	}

	// The middle of an even length is its own mirror image.
	var mr, mg, mb float64
	if n%2 == 0 {
		mr, mg, mb = v[n/2][0], v[n/2][1], v[n/2][2]
		dr, dg, db = dr+mr, dg+mg, db+mb
	}

	out := e.result[:len(cos)]
	out[0] = [3]float64{dr, dg, db}

	for k := 1; k < len(cos); k++ {
		// Even k take the pair sums and odd k the differences.
		pairs := sum
		if k%2 == 1 {
			pairs = dif
		}

		r, g, b := dot(cos[k][1:m+1], pairs)
		r, g, b = fr+r, fg+g, fb+b

		if n%2 == 0 {
			c := cos[k][n/2]
			r += float64(c * mr)
			g += float64(c * mg)
			b += float64(c * mb)
		}

		out[k] = [3]float64{r, g, b}
	}

	return out
}

// dot returns the sum of c[k] times v[k] for each channel.
func dot(c []float64, v [][3]float64) (r, g, b float64) {
	v = v[:len(c)]

	for k, ck := range c {
		// Each product is rounded before it is added, so no target fuses the two.
		r += float64(ck * v[k][0])
		g += float64(ck * v[k][1])
		b += float64(ck * v[k][2])
	}

	return r, g, b
}

// grow returns s resliced to n, reallocated if it is too short.
func grow(s [][3]float64, n int) [][3]float64 {
	if cap(s) < n {
		return make([][3]float64, n)
	}

	return s[:n]
}
