package blurhash

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"image"
	"math"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

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

// TestEncodeSize requires Encode and Encoder.Encode to return
// ErrInvalidDimensions for an image with no pixels even when its other side is
// too large, ErrImageTooLarge past MaxSize or for a side that does not fit in
// an int, and no error at MaxSize.
func TestEncodeSize(t *testing.T) {
	var e Encoder

	// A Rectangle is an image with any bounds.
	for _, c := range []struct {
		r    image.Rectangle
		want error
	}{
		{image.Rect(0, 0, 0, MaxSize+1), ErrInvalidDimensions},
		{image.Rect(0, 0, MaxSize+1, 1), ErrImageTooLarge},
		{image.Rect(0, 0, 1, MaxSize+1), ErrImageTooLarge},
		{image.Rect(-2, 0, math.MaxInt, 1), ErrImageTooLarge},
		{image.Rect(0, math.MinInt, 1, 1), ErrImageTooLarge},
		{image.Rect(0, 0, MaxSize, 1), nil},
		{image.Rect(0, 0, 1, MaxSize), nil},
	} {
		if _, err := Encode(4, 3, c.r); !errors.Is(err, c.want) {
			t.Errorf("Encode of %v: error %v, want %v", c.r, err, c.want)
		}

		if _, err := e.Encode(4, 3, c.r); !errors.Is(err, c.want) {
			t.Errorf("Encoder.Encode of %v: error %v, want %v", c.r, err, c.want)
		}
	}
}

// TestEncodeComparison requires Encode and Encoder.Encode to agree on every
// corpus hash, and the hashes that differ from comparison.txt to be exactly
// those in testdata/differences.txt.
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

// TestEncodeYCbCrPanics requires Encode to panic wherever At panics on a
// YCbCr image, never reading a byte outside it: on the chroma planes NewYCbCr
// makes too small, and on planes one byte short with that byte still in
// capacity.
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

// FuzzEncode requires Encode to accept exactly 1 to 9 components on each axis
// and images 1 to MaxSize pixels on each side, and to return a hash that
// Components and Decode accept.
func FuzzEncode(f *testing.F) {
	f.Add(int16(0), int16(0), int16(8), int16(6), 4, 3, []byte{1, 2, 3, 255})
	f.Add(int16(-3), int16(5), int16(1), int16(1), 1, 1, []byte{200})
	f.Add(int16(0), int16(0), int16(0), int16(5), 4, 3, []byte{0})
	f.Add(int16(0), int16(0), int16(MaxSize+1), int16(1), 4, 3, []byte{0})
	f.Add(int16(0), int16(0), int16(4), int16(4), 0, 10, []byte{0})

	f.Fuzz(func(t *testing.T, x0, y0, w, h int16, nx, ny int, pix []byte) {
		r := image.Rectangle{Min: image.Pt(int(x0), int(y0)), Max: image.Pt(int(x0)+int(w), int(y0)+int(h))}

		sized := !r.Empty() && r.Dx() <= MaxSize && r.Dy() <= MaxSize
		if sized && r.Dx()*r.Dy() > 1<<16 {
			return // the large accepted sizes are TestEncodeSize's rows
		}

		// The fuzzed bytes, repeated to fill the image; an image Encode must
		// refuse is the bare rectangle.
		var img image.Image = r
		if sized {
			buf := make([]byte, 4*r.Dx()*r.Dy())
			if len(pix) > 0 {
				for i := range buf {
					buf[i] = pix[i%len(pix)]
				}
			}

			img = &image.NRGBA{Pix: buf, Stride: 4 * r.Dx(), Rect: r}
		}

		hash, err := Encode(nx, ny, img)
		if ok := sized && nx >= 1 && nx <= 9 && ny >= 1 && ny <= 9; (err == nil) != ok {
			t.Fatalf("%v at %dx%d: %v", r, nx, ny, err)
		}

		if err != nil {
			return
		}

		if x, y, err := Components(hash); x != nx || y != ny || err != nil {
			t.Fatalf("%v at %dx%d: %q has components %d, %d, %v", r, nx, ny, hash, x, y, err)
		}

		if _, err := Decode(hash, 4, 4, 1); err != nil {
			t.Fatalf("%v at %dx%d: %q does not decode: %v", r, nx, ny, hash, err)
		}
	})
}

var update = flag.Bool("update", false, "rewrite the files in testdata from the current encoder")

// testdata is embedded so the tests run where only the test binary is.
//
//go:embed testdata
var testdata embed.FS

func panics(f func()) (p bool) {
	defer func() { p = recover() != nil }()
	f()

	return false
}
