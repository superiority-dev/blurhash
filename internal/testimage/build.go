package testimage

import (
	"image"
	"image/color"
)

// The offset family tests images whose bounds do not start at (0, 0). Each is
// cut from a bigger noise image: it starts at offsetMin, and the bigger image
// extends offsetMargin past its right and bottom edges.
var offsetMin, offsetMargin = image.Pt(5, 3), image.Pt(4, 4)

func solid(s image.Point, v uint8) func() image.Image {
	return func() image.Image { return fill(s, func(int, int) color.RGBA { return gray(v) }) }
}

// splitX gives the left half of s gray a and the right half gray b.
func splitX(s image.Point, a, b uint8) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, _ int) color.RGBA {
			if x < s.X/2 {
				return gray(a)
			}

			return gray(b)
		})
	}
}

// splitY gives the top half of s gray a and the bottom half gray b.
func splitY(s image.Point, a, b uint8) func() image.Image {
	return func() image.Image {
		return fill(s, func(_, y int) color.RGBA {
			if y < s.Y/2 {
				return gray(a)
			}

			return gray(b)
		})
	}
}

// whiteRun makes the first n pixels of the top row white and the rest black.
func whiteRun(s image.Point, n int) func() image.Image {
	return func() image.Image {
		return fill(s, func(x, y int) color.RGBA {
			if y == 0 && x < n {
				return gray(255)
			}

			return gray(0)
		})
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
func typed(kind imageKind, s image.Point) func() image.Image {
	return func() image.Image { return convert(kind, fill(s, noise(uint64(s.X*s.Y)))) }
}

func translucent(kind imageKind, s image.Point) func() image.Image {
	return func() image.Image {
		r := newSplitMix(alphaSeed)
		b := image.Rectangle{Max: s}

		switch kind {
		case kindNRGBA:
			img := image.NewNRGBA(b)
			for i := range img.Pix {
				img.Pix[i] = r.byte()
			}

			return img
		case kindRGBA:
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

		panic("testimage: translucent images are rgba or nrgba")
	}
}

// offset builds an opaque s-sized noise sub-image of the given kind.
func offset(kind imageKind, s image.Point) func() image.Image {
	return func() image.Image {
		sub := image.Rectangle{Min: offsetMin, Max: offsetMin.Add(s)}
		full := convert(kind, fill(sub.Max.Add(offsetMargin), noise(offsetSeed)))

		return full.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(sub)
	}
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

func gray(v uint8) color.RGBA {
	return color.RGBA{R: v, G: v, B: v, A: 255}
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
