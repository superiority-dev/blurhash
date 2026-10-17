package blurhash

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"image"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

var update = flag.Bool("update", false, "rewrite the files in testdata from the current encoder")

// testdata is embedded so the tests run where only the test binary is.
//
//go:embed testdata
var testdata embed.FS

func TestEncodeErrors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))

	// Each argument just below 1 and just above 9, and a negative.
	for _, c := range []image.Point{{0, 1}, {1, 0}, {10, 1}, {1, 10}, {-1, 3}} {
		if _, err := Encode(c.X, c.Y, img); !errors.Is(err, ErrInvalidComponents) {
			t.Errorf("Encode(%d, %d): error %v, want ErrInvalidComponents", c.X, c.Y, err)
		}
	}

	// No pixels.
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, 0, 0), // empty
		image.Rect(0, 0, 5, 0), // zero height
		image.Rect(3, 3, 3, 9), // zero width, away from the origin
	} {
		if _, err := Encode(4, 3, image.NewRGBA(r)); !errors.Is(err, ErrInvalidDimensions) {
			t.Errorf("Encode of %v: error %v, want ErrInvalidDimensions", r, err)
		}
	}
}

// Every hash that differs from comparison.txt must be listed in
// testdata/differences.txt, so a new difference and a vanished one both fail.
func TestEncodeComparison(t *testing.T) {
	file, err := testimage.Comparisons()
	if err != nil {
		t.Fatal(err)
	}

	var (
		got []string
		e   Encoder
	)

	n := 0

	for _, c := range testimage.Cases() {
		img := c.Image()

		for _, comp := range c.Components {
			h, err := Encode(comp.X, comp.Y, img)
			if err != nil {
				t.Fatalf("%s at %v: %v", c.ID, comp, err)
			}

			if eh, _ := e.Encode(comp.X, comp.Y, img); eh != h {
				t.Fatalf("%s at %v: Encoder.Encode %s, Encode %s", c.ID, comp, eh, h)
			}

			if h != file.Results[n].Hash {
				got = append(got, fmt.Sprintf("%s %v %s", c.ID, comp, h))
			}

			n++
		}
	}

	testimage.Golden(t, testdata, "testdata/differences.txt", got, *update)
}

func TestEncodeAllocs(t *testing.T) {
	// One image for each pixel format Encode reads directly.
	images := []image.Image{
		testimage.Random(1, 32, 24, 256),
		image.NewRGBA(image.Rect(0, 0, 32, 24)),
		image.NewYCbCr(image.Rect(0, 0, 32, 24), image.YCbCrSubsampleRatio420),
	}

	for _, img := range images {
		var e Encoder

		if n := testing.AllocsPerRun(100, func() { _, _ = e.Encode(4, 4, img) }); n != 1 {
			t.Errorf("%T: Encoder.Encode allocates %v times, want 1", img, n)
		}

		if raceEnabled {
			continue
		}

		if n := testing.AllocsPerRun(100, func() { _, _ = Encode(4, 4, img) }); n != 1 {
			t.Errorf("%T: Encode allocates %v times, want 1", img, n)
		}
	}
}

// Encode reads YCbCr planes directly, so it must panic wherever At does
// instead of reading a byte outside the image: on the chroma planes
// NewYCbCr makes too small, and on planes one byte short with that byte
// still in capacity.
func TestEncodeYCbCrPanics(t *testing.T) {
	shortY := image.NewYCbCr(image.Rect(0, 0, 8, 4), image.YCbCrSubsampleRatio444)
	shortY.Y = shortY.Y[:len(shortY.Y)-1]

	shortCr := image.NewYCbCr(image.Rect(0, 0, 8, 4), image.YCbCrSubsampleRatio420)
	shortCr.Cr = shortCr.Cr[:len(shortCr.Cr)-1]

	for _, img := range []*image.YCbCr{
		// Odd negative origins, where NewYCbCr's chroma planes come out too small.
		image.NewYCbCr(image.Rect(-9, -9, 0, -8), image.YCbCrSubsampleRatio422),
		image.NewYCbCr(image.Rect(-9, -9, -8, 0), image.YCbCrSubsampleRatio420),
		image.NewYCbCr(image.Rect(-9, -9, -8, 0), image.YCbCrSubsampleRatio440),
		image.NewYCbCr(image.Rect(-9, -9, -2, -8), image.YCbCrSubsampleRatio411),
		image.NewYCbCr(image.Rect(-9, -9, -8, 0), image.YCbCrSubsampleRatio410),
		shortY,
		shortCr,
	} {
		at := panics(func() {
			for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
				for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
					img.At(x, y)
				}
			}
		})

		if enc := panics(func() { _, _ = Encode(1, 1, img) }); !at || !enc {
			t.Errorf("%v %v: At panics %v, Encode panics %v; want both",
				img.SubsampleRatio, img.Rect, at, enc)
		}
	}
}

func panics(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()

	return false
}
