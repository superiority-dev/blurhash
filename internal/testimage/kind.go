package testimage

import (
	"image"
	"image/color"
	"image/draw"
)

const paletteSize = 64

// kinds lists the kinds of the type family, in case order.
var kinds = []imageKind{kindNRGBA, kindGray, kindYCbCr, kindPaletted, kindRGBA64, kindCMYK}

// imageKind is the type a test image is stored as: one of package image's
// pixel-buffer types. Its name appears in case IDs, as in type/gray/33x41.
type imageKind uint8

const (
	kindNRGBA imageKind = iota
	kindGray
	kindYCbCr
	kindPaletted
	kindRGBA64
	kindCMYK
	kindRGBA
)

func (k imageKind) String() string {
	return [...]string{"nrgba", "gray", "ycbcr", "paletted", "rgba64", "cmyk", "rgba"}[k]
}

// convert stores src in an image of the given kind.
func convert(kind imageKind, src *image.RGBA) image.Image {
	b := src.Bounds()

	switch kind {
	case kindNRGBA:
		dst := image.NewNRGBA(b)
		copy(dst.Pix, src.Pix)

		return dst
	case kindGray:
		return drawInto(image.NewGray(b), src)
	case kindYCbCr:
		dst := image.NewYCbCr(b, image.YCbCrSubsampleRatio444)

		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := src.RGBAAt(x, y)
				yy, cb, cr := color.RGBToYCbCr(c.R, c.G, c.B)
				dst.Y[dst.YOffset(x, y)] = yy
				dst.Cb[dst.COffset(x, y)] = cb
				dst.Cr[dst.COffset(x, y)] = cr
			}
		}

		return dst
	case kindPaletted:
		return drawInto(image.NewPaletted(b, palette()), src)
	case kindRGBA64:
		return drawInto(image.NewRGBA64(b), src)
	case kindCMYK:
		return drawInto(image.NewCMYK(b), src)
	case kindRGBA:
		return src
	}

	panic("testimage: unknown image kind")
}

// drawInto copies src into dst through dst's color model.
func drawInto(dst draw.Image, src *image.RGBA) image.Image {
	b := src.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}

	return dst
}

func palette() color.Palette {
	r := newSplitMix(paletteSeed)
	p := make(color.Palette, paletteSize)

	for i := range p {
		p[i] = color.RGBA{R: r.byte(), G: r.byte(), B: r.byte(), A: 255}
	}

	return p
}
