package codec

import (
	"image"

	"github.com/superiority-dev/blurhash/internal/ycbcr"
)

// ties collects, per channel, what dcTie needs: the sum of all bytes, the
// count of 255s, and whether any byte lay strictly between 10 and 255.
type ties struct {
	sum, white [3]uint64
	other      [3]bool
}

// read converts row y of img to linear light in line, which has the image's
// width, and adds the row's bytes to t.
func (t *ties) read(img image.Image, y int, line [][3]float64) {
	b := img.Bounds()

	var sr, sg, sb, wr, wg, wb uint64

	var or, og, ob uint8

	add := func(x int, r, g, bl uint8) {
		line[x] = [3]float64{linear[r], linear[g], linear[bl]}
		sr, sg, sb = sr+uint64(r), sg+uint64(g), sb+uint64(bl)
		wr, wg, wb = wr+white(r), wg+white(g), wb+white(bl)
		or, og, ob = or|middle(r), og|middle(g), ob|middle(bl)
	}

	switch img := img.(type) {
	case *image.RGBA:
		pix := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		pix = pix[:4*len(line)]

		for x := range line {
			add(x, pix[4*x], pix[4*x+1], pix[4*x+2])
		}
	case *image.NRGBA:
		pix := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		pix = pix[:4*len(line)]

		for x := range line {
			if a := uint32(pix[4*x+3]); a != 0xff {
				add(x, premultiply(pix[4*x], a), premultiply(pix[4*x+1], a), premultiply(pix[4*x+2], a))
			} else {
				add(x, pix[4*x], pix[4*x+1], pix[4*x+2])
			}
		}
	case *image.YCbCr:
		// Read the planes by index, with COffset's column arithmetic.
		x0 := b.Min.X
		ys := img.Y[img.YOffset(x0, b.Min.Y+y):]
		_ = ys[len(line)-1] // At checks the plane's length; slicing checks only its capacity
		ys = ys[:len(line)]
		c := img.COffset(x0, b.Min.Y+y)
		cb, cr := img.Cb[c:], img.Cr[c:]

		switch img.SubsampleRatio {
		case image.YCbCrSubsampleRatio422, image.YCbCrSubsampleRatio420:
			for x, v := range ys {
				i := (x0+x)/2 - x0/2
				r, g, bl := ycbcr.RGB(v, cb[i], cr[i])
				add(x, r, g, bl)
			}
		case image.YCbCrSubsampleRatio411, image.YCbCrSubsampleRatio410:
			for x, v := range ys {
				i := (x0+x)/4 - x0/4
				r, g, bl := ycbcr.RGB(v, cb[i], cr[i])
				add(x, r, g, bl)
			}
		default:
			for x, v := range ys {
				r, g, bl := ycbcr.RGB(v, cb[x], cr[x])
				add(x, r, g, bl)
			}
		}
	default:
		for x := range line {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			add(x, uint8(r>>8), uint8(g>>8), uint8(bl>>8))
		}
	}

	t.sum[0], t.sum[1], t.sum[2] = t.sum[0]+sr, t.sum[1]+sg, t.sum[2]+sb
	t.white[0], t.white[1], t.white[2] = t.white[0]+wr, t.white[1]+wg, t.white[2]+wb
	t.other[0], t.other[1], t.other[2] = t.other[0] || or != 0, t.other[1] || og != 0, t.other[2] || ob != 0
}

// dc returns each channel's DC tie, as dcTie does, or -1, for an image of n
// pixels.
func (t *ties) dc(n int) [3]int {
	out := [3]int{-1, -1, -1}

	for c := range out {
		if !t.other[c] {
			out[c] = dcTie(t.sum[c]-255*t.white[c], t.white[c], uint64(n))
		}
	}

	return out
}

// premultiply returns the 8-bit value color.NRGBA's RGBA method gives for
// channel v at alpha a.
func premultiply(v uint8, a uint32) uint8 {
	return uint8(uint32(v) * 0x101 * a / 0xff >> 8)
}

// white returns 1 for 255 and 0 otherwise.
func white(v uint8) uint64 {
	return (uint64(v) + 1) >> 8
}

// middle returns 1 for a byte strictly between 10 and 255 and 0 otherwise.
func middle(v uint8) uint8 {
	if v-11 < 244 {
		return 1
	}

	return 0
}
