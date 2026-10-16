package testimage

import (
	"image"
	"image/color"
)

var kinds = []string{"nrgba", "gray", "ycbcr", "paletted", "rgba64", "cmyk", "custom"}

type setter interface {
	image.Image
	Set(x, y int, c color.Color)
}

// custom hides an RGBA behind a type the encoder has no fast path for.
type custom struct {
	img *image.RGBA
}

func (c custom) ColorModel() color.Model { return color.RGBAModel }
func (c custom) Bounds() image.Rectangle { return c.img.Bounds() }
func (c custom) At(x, y int) color.Color { return c.img.At(x, y) }

func solid(s image.Point, v uint8) func() image.Image {
	return func() image.Image { return fill(s, func(int, int) color.RGBA { return grey(v) }) }
}

// splitX gives the left half of s grey a and the right half grey b.
func splitX(s image.Point, a, b uint8) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, _ int) color.RGBA { return grey(pick(x < s.X/2, a, b)) })
	}
}

// splitY gives the top half of s grey a and the bottom half grey b.
func splitY(s image.Point, a, b uint8) func() image.Image {
	return func() image.Image {
		return fill(s, func(_, y int) color.RGBA { return grey(pick(y < s.Y/2, a, b)) })
	}
}

// whiteRun makes the first n pixels of the top row white and the rest black.
func whiteRun(s image.Point, n int) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, y int) color.RGBA { return grey(pick(y == 0 && x < n, 255, 0)) })
	}
}

// gradient ramps red across, green down and blue along the diagonal.
func gradient(s image.Point) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, y int) color.RGBA {
			return color.RGBA{R: ramp(x, s.X), G: ramp(y, s.Y), B: ramp(x+y, s.X+s.Y-1), A: 255}
		})
	}
}

func smooth(s image.Point) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, y int) color.RGBA {
			return color.RGBA{R: bump(x, s.X), G: bump(y, s.Y), B: bump(x+y, s.X+s.Y-1), A: 255}
		})
	}
}

func noisy(s image.Point, seed uint64) func() image.Image {
	return func() image.Image { return fill(s, noise(seed)) }
}

// typed is noise seeded by the size of s, in image type kind.
func typed(kind string, s image.Point) func() image.Image {
	return func() image.Image { return convert(kind, fill(s, noise(uint64(s.X*s.Y)))) }
}

func translucent(kind string, s image.Point) func() image.Image {
	return func() image.Image {
		r := newSplitMix(0xa1fa)
		b := image.Rectangle{Max: s}

		if kind == "nrgba" {
			img := image.NewNRGBA(b)
			for i := range img.Pix {
				img.Pix[i] = r.byte()
			}

			return img
		}

		img := image.NewRGBA(b)

		for i := 0; i < len(img.Pix); i += 4 {
			a := uint16(r.byte())
			for c := range 3 {
				img.Pix[i+c] = uint8(uint16(r.byte()) * a / 255)
			}

			img.Pix[i+3] = uint8(a)
		}

		return img
	}
}

// offset returns an opaque s sub-image whose bounds start at (5, 3).
func offset(kind string, s image.Point) func() image.Image {
	return func() image.Image {
		full := fill(image.Pt(s.X+9, s.Y+7), noise(0x0ff5e7))
		sub := image.Rect(5, 3, 5+s.X, 3+s.Y)

		if kind == "nrgba" {
			return convert("nrgba", full).(*image.NRGBA).SubImage(sub)
		}

		return full.SubImage(sub)
	}
}

func convert(kind string, src *image.RGBA) image.Image {
	b := src.Bounds()

	switch kind {
	case "nrgba":
		dst := image.NewNRGBA(b)
		copy(dst.Pix, src.Pix)

		return dst
	case "gray":
		return drawInto(image.NewGray(b), src)
	case "ycbcr":
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
	case "paletted":
		return drawInto(image.NewPaletted(b, palette()), src)
	case "rgba64":
		return drawInto(image.NewRGBA64(b), src)
	case "cmyk":
		return drawInto(image.NewCMYK(b), src)
	case "custom":
		return custom{src}
	}

	panic("testimage: unknown image kind " + kind)
}

// drawInto copies src into dst through dst's colour model.
func drawInto(dst setter, src *image.RGBA) image.Image {
	b := src.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}

	return dst
}

func palette() color.Palette {
	r := newSplitMix(0xb1a5)
	p := make(color.Palette, 64)

	for i := range p {
		p[i] = color.RGBA{R: r.byte(), G: r.byte(), B: r.byte(), A: 255}
	}

	return p
}

func fill(s image.Point, at func(x, y int) color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{Max: s})

	for y := range s.Y {
		for x := range s.X {
			img.SetRGBA(x, y, at(x, y))
		}
	}

	return img
}

func noise(seed uint64) func(int, int) color.RGBA {
	r := newSplitMix(seed)

	return func(int, int) color.RGBA {
		return color.RGBA{R: r.byte(), G: r.byte(), B: r.byte(), A: 255}
	}
}

func grey(v uint8) color.RGBA {
	return color.RGBA{R: v, G: v, B: v, A: 255}
}

func pick(first bool, a, b uint8) uint8 {
	if first {
		return a
	}

	return b
}

// bump maps i in [0, n) onto a parabola rising from 0 to 255 and back.
func bump(i, n int) uint8 {
	return uint8(4 * 255 * i * (n - 1 - i) / ((n - 1) * (n - 1)))
}

func ramp(i, n int) uint8 {
	if n <= 1 {
		return 0
	}

	return uint8(i * 255 / (n - 1))
}
