package codec

import (
	"image"

	"github.com/superiority-dev/blurhash/internal/ycbcr"
)

// linearRow writes row y of img, in linear light, into dst, which has the
// image's width.
func linearRow(dst [][3]float64, img image.Image, y int) {
	b := img.Bounds()

	switch img := img.(type) {
	case *image.RGBA:
		pix := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		pix = pix[:4*len(dst)]

		for x := range dst {
			r, g, bl := pix[4*x], pix[4*x+1], pix[4*x+2]
			dst[x] = [3]float64{linear[r], linear[g], linear[bl]}
		}
	case *image.NRGBA:
		pix := img.Pix[img.PixOffset(b.Min.X, b.Min.Y+y):]
		pix = pix[:4*len(dst)]

		for x := range dst {
			r, g, bl := pix[4*x], pix[4*x+1], pix[4*x+2]
			if a := uint32(pix[4*x+3]); a != 0xff {
				r, g, bl = premultiply(r, a), premultiply(g, a), premultiply(bl, a)
			}

			dst[x] = [3]float64{linear[r], linear[g], linear[bl]}
		}
	case *image.YCbCr:
		// Read the planes by index, with COffset's column arithmetic.
		x0 := b.Min.X
		ys := img.Y[img.YOffset(x0, b.Min.Y+y):]
		_ = ys[len(dst)-1] // At checks the plane's length; slicing checks only its capacity
		ys = ys[:len(dst)]
		c := img.COffset(x0, b.Min.Y+y)
		cb, cr := img.Cb[c:], img.Cr[c:]

		switch img.SubsampleRatio {
		case image.YCbCrSubsampleRatio422, image.YCbCrSubsampleRatio420:
			for x, v := range ys {
				i := (x0+x)/2 - x0/2
				r, g, bl := ycbcr.RGB(v, cb[i], cr[i])
				dst[x] = [3]float64{linear[r], linear[g], linear[bl]}
			}
		case image.YCbCrSubsampleRatio411, image.YCbCrSubsampleRatio410:
			for x, v := range ys {
				i := (x0+x)/4 - x0/4
				r, g, bl := ycbcr.RGB(v, cb[i], cr[i])
				dst[x] = [3]float64{linear[r], linear[g], linear[bl]}
			}
		default:
			for x, v := range ys {
				r, g, bl := ycbcr.RGB(v, cb[x], cr[x])
				dst[x] = [3]float64{linear[r], linear[g], linear[bl]}
			}
		}
	default:
		for x := range dst {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			dst[x] = [3]float64{linear[r>>8], linear[g>>8], linear[bl>>8]}
		}
	}
}

// premultiply returns the 8-bit value color.NRGBA's RGBA method gives for
// channel v at alpha a.
func premultiply(v uint8, a uint32) uint8 {
	return uint8(uint32(v) * 0x101 * a / 0xff >> 8)
}
