// Package testimage builds the test corpus and holds its expected hashes.
// Each image is a pure function of its ID.
package testimage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"slices"
)

// Components counts a hash's horizontal (X) and vertical (Y) components.
type Components struct {
	X, Y int
}

// String returns c as XxY, such as 4x3.
func (c Components) String() string {
	return fmt.Sprintf("%dx%d", c.X, c.Y)
}

// Case is one corpus entry.
type Case struct {
	ID         string
	Components []Components
	build      func() image.Image
}

// Image returns a new copy of the case's image on every call.
func (c Case) Image() image.Image {
	return c.build()
}

var (
	small = []Components{
		{1, 1}, {1, 2}, {2, 1}, {3, 3}, {4, 3}, {3, 4}, {5, 5}, {9, 1}, {1, 9}, {9, 9},
	}
	large = []Components{
		{1, 1}, {4, 3}, {9, 9},
	}

	// Smallest, even and odd.
	greySizes = []image.Point{{1, 1}, {8, 8}, {9, 7}}

	// Smallest, even, odd, and the 100x75 thumbnail.
	splitSizes = []image.Point{{2, 2}, {8, 8}, {9, 7}, {100, 75}}

	gradientSizes = []image.Point{{255, 1}, {1, 255}, {17, 9}, {100, 75}, {301, 97}, {1000, 750}}

	// From 1x1 to 1000x750: odd and even, square, wide and tall.
	noiseSizes = []image.Point{
		{1, 1}, {1, 7}, {7, 1}, {2, 3}, {8, 8}, {9, 9}, {13, 7}, {17, 31},
		{32, 32}, {33, 41}, {64, 48}, {97, 1}, {100, 75}, {101, 76}, {127, 83},
		{160, 90}, {255, 1}, {256, 171}, {301, 97}, {333, 333}, {400, 300},
		{640, 480}, {999, 749}, {1000, 750},
	}

	// Noise sizes also encoded at 4x4.
	fourByFour = map[image.Point]bool{{32, 32}: true, {100, 75}: true}

	// What 32 and 100 pixels on the long edge give at 4:3, 3:2 and 16:9.
	thumbSizes = []image.Point{
		{32, 24}, {24, 32}, {32, 21}, {21, 32}, {32, 18}, {18, 32},
		{75, 100}, {100, 66}, {66, 100}, {100, 56}, {56, 100},
	}

	typeSizes = []image.Point{{33, 41}, {100, 75}}
)

// Cases returns the corpus in the order comparison.txt lists it.
func Cases() []Case {
	var out []Case

	add := func(id string, s image.Point, build func() image.Image) {
		comps := small
		if s.X*s.Y > 400*300 {
			comps = large
		}

		out = append(out, Case{ID: id, Components: comps, build: build})
	}

	for _, s := range greySizes {
		for v := range 256 {
			add(fmt.Sprintf("grey/%d/%v", v, size(s)), s, solid(s, uint8(v)))
		}
	}

	pairs := [][2]uint8{{0, 255}, {127, 128}, {254, 255}}
	// Dark pairs 0/1 to 15/16, across where the sRGB curve stops being linear (byte 10).
	for v := range 16 {
		pairs = append(pairs, [2]uint8{uint8(v), uint8(v + 1)})
	}

	for _, s := range splitSizes {
		for _, p := range pairs {
			add(fmt.Sprintf("split-x/%d-%d/%v", p[0], p[1], size(s)), s, splitX(s, p[0], p[1]))
			add(fmt.Sprintf("split-y/%d-%d/%v", p[0], p[1], size(s)), s, splitY(s, p[0], p[1]))
		}
	}

	// Black and white splits whose exact ties expose one ulp of cos or a fused multiply-add.
	add("split-x/0-255/18x8", image.Pt(18, 8), splitX(image.Pt(18, 8), 0, 255))
	add("split-y/0-255/72x18", image.Pt(72, 18), splitY(image.Pt(72, 18), 0, 255))
	add("split-x/0-255/2x8", image.Pt(2, 8), splitX(image.Pt(2, 8), 0, 255))
	add("split-y/0-255/8x2", image.Pt(8, 2), splitY(image.Pt(8, 2), 0, 255))

	// A DC that is an exact tie.
	add("white-dc/38x51", image.Pt(38, 51), whiteRun(image.Pt(38, 51), 5))

	for _, s := range gradientSizes {
		add("gradient/"+size(s), s, gradient(s))
	}

	add("smooth/1000x750", image.Pt(1000, 750), smooth(image.Pt(1000, 750)))

	for i, s := range noiseSizes {
		add(fmt.Sprintf("noise/%d/%v", i, size(s)), s, noisy(s, uint64(i)))

		if fourByFour[s] {
			out[len(out)-1].Components = append(slices.Clip(out[len(out)-1].Components), Components{4, 4})
		}
	}

	for i, s := range thumbSizes {
		seed := uint64(24 + i)
		out = append(out, Case{
			ID:         fmt.Sprintf("noise/%d/%v", seed, size(s)),
			Components: []Components{{4, 4}, {4, 3}},
			build:      noisy(s, seed),
		})
	}

	// An AC value 2e-9 from a quantisation boundary, not a tie.
	add("near/26318/3x3", image.Pt(3, 3), noisy(image.Pt(3, 3), 26318))

	for _, s := range typeSizes {
		for _, kind := range kinds {
			add(fmt.Sprintf("type/%s/%v", kind, size(s)), s, typed(kind, s))
		}
	}

	for _, kind := range []string{"rgba", "nrgba"} {
		add("alpha/"+kind+"/16x16", image.Pt(16, 16), translucent(kind, image.Pt(16, 16)))
		add("offset/"+kind+"/16x16", image.Pt(16, 16), offset(kind, image.Pt(16, 16)))
	}

	return out
}

// Fingerprint returns the SHA-256 of every case ID, component list and pixel.
func Fingerprint() string {
	h := sha256.New()

	for _, c := range Cases() {
		fmt.Fprintf(h, "%s %v\n", c.ID, c.Components)

		for _, p := range RGBA(c.Image()) {
			h.Write(p[:])
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// RGBA returns img's pixels in row order as 8-bit premultiplied RGBA.
func RGBA(img image.Image) [][4]uint8 {
	b := img.Bounds()
	out := make([][4]uint8, 0, b.Dx()*b.Dy())

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			out = append(out, [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)})
		}
	}

	return out
}

func size(p image.Point) string {
	return fmt.Sprintf("%dx%d", p.X, p.Y)
}
