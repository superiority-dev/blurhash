package blurhash

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"math"
	"slices"
	"testing"

	"github.com/superiority-dev/blurhash/internal/codec"
)

const (
	hash4x3 = "LEHV6nWB2yk8pyo0adR*.7kCMdnj"
	hash1x1 = "000000"
)

func TestComponents(t *testing.T) {
	if x, y, err := Components(hash4x3); x != 4 || y != 3 || err != nil {
		t.Errorf("Components: %d, %d, %v; want 4, 3, nil", x, y, err)
	}
}

func TestDecodeInvalidHash(t *testing.T) {
	for _, hash := range []string{
		"",
		"00000",      // shorter than any hash
		"!0000000",   // not a base-83 digit
		"}000000000", // size 81 would mean 10 vertical components
		"000000000",  // 1x1 needs exactly 6 characters
		"1000000",    // 2x1 needs 8 characters: one short
		"0!0000",     // maximum is not a digit
		"00~~~~",     // average color above 0xffffff
		"100000~~",   // 2x1 whose AC holds a digit above 18
		"100000!0",   // AC is not a digit
	} {
		if _, err := Decode(hash, 4, 4, 1); !errors.Is(err, ErrInvalidHash) {
			t.Errorf("Decode(%q): error %v, want ErrInvalidHash", hash, err)
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	// No pixels: zero width, zero height and negative width.
	for _, s := range []image.Point{{0, 1}, {1, 0}, {-3, 4}} {
		if _, err := Decode(hash1x1, s.X, s.Y, 1); !errors.Is(err, ErrInvalidDimensions) {
			t.Errorf("Decode at %dx%d: error %v, want ErrInvalidDimensions", s.X, s.Y, err)
		}
	}

	// An empty destination is not an error and is left alone.
	if err := DecodeDraw(image.NewNRGBA(image.Rectangle{}), hash1x1, 1); err != nil {
		t.Errorf("DecodeDraw into an empty image: %v", err)
	}
}

// TestDecodeSize requires Decode and Decoder.Decode to return
// ErrInvalidDimensions for a size with no pixels even when its other side is
// too large, ErrImageTooLarge past MaxSize, and no error at MaxSize;
// DecodeDraw and Decoder.DecodeDraw the same limit on any bounds; and a
// Decoder's MaxSize field to replace that limit, with zero or less keeping it.
func TestDecodeSize(t *testing.T) {
	var d Decoder

	for _, c := range []struct {
		s    image.Point
		want error
	}{
		{image.Pt(0, MaxSize+1), ErrInvalidDimensions},
		{image.Pt(MaxSize+1, 1), ErrImageTooLarge},
		{image.Pt(1, MaxSize+1), ErrImageTooLarge},
		{image.Pt(math.MaxInt, math.MaxInt), ErrImageTooLarge},
		{image.Pt(MaxSize, 1), nil},
		{image.Pt(1, MaxSize), nil},
	} {
		if _, err := Decode(hash1x1, c.s.X, c.s.Y, 1); !errors.Is(err, c.want) {
			t.Errorf("Decode at %v: error %v, want %v", c.s, err, c.want)
		}

		if _, err := d.Decode(hash1x1, c.s.X, c.s.Y, 1); !errors.Is(err, c.want) {
			t.Errorf("Decoder.Decode at %v: error %v, want %v", c.s, err, c.want)
		}
	}

	for _, c := range []struct {
		r    image.Rectangle
		want error
	}{
		{image.Rect(0, 0, MaxSize+1, 1), ErrImageTooLarge},
		{image.Rect(0, 0, 1, MaxSize+1), ErrImageTooLarge},
		{image.Rect(-2, 0, math.MaxInt, 1), ErrImageTooLarge},
		{image.Rect(0, math.MinInt, 1, 1), ErrImageTooLarge},
		{image.Rect(0, 0, MaxSize, 1), nil},
		{image.Rect(0, 0, 1, MaxSize), nil},
	} {
		if err := DecodeDraw(wide{c.r}, hash1x1, 1); !errors.Is(err, c.want) {
			t.Errorf("DecodeDraw into %v: error %v, want %v", c.r, err, c.want)
		}

		if err := d.DecodeDraw(wide{c.r}, hash1x1, 1); !errors.Is(err, c.want) {
			t.Errorf("Decoder.DecodeDraw into %v: error %v, want %v", c.r, err, c.want)
		}
	}

	for _, c := range []struct {
		max  int
		s    image.Point
		want error
	}{
		{8, image.Pt(8, 8), nil},
		{8, image.Pt(9, 1), ErrImageTooLarge},
		{8, image.Pt(1, 9), ErrImageTooLarge},
		{MaxSize + 1, image.Pt(MaxSize+1, 1), nil},
		{-1, image.Pt(MaxSize, 1), nil},
		{-1, image.Pt(MaxSize+1, 1), ErrImageTooLarge},
	} {
		d := Decoder{MaxSize: c.max}
		if _, err := d.Decode(hash1x1, c.s.X, c.s.Y, 1); !errors.Is(err, c.want) {
			t.Errorf("Decoder{MaxSize: %d}.Decode at %v: error %v, want %v", c.max, c.s, err, c.want)
		}

		if err := d.DecodeDraw(wide{image.Rectangle{Max: c.s}}, hash1x1, 1); !errors.Is(err, c.want) {
			t.Errorf("Decoder{MaxSize: %d}.DecodeDraw into %v: error %v, want %v", c.max, c.s, err, c.want)
		}
	}
}

func TestDecodeNonFinitePunch(t *testing.T) {
	for _, punch := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := Decode(hash4x3, 8, 8, punch); !errors.Is(err, ErrInvalidPunch) {
			t.Errorf("Decode at punch %v: error %v, want ErrInvalidPunch", punch, err)
		}

		dst := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		if err := DecodeDraw(dst, hash4x3, punch); !errors.Is(err, ErrInvalidPunch) {
			t.Errorf("DecodeDraw at punch %v: error %v, want ErrInvalidPunch", punch, err)
		}

		if slices.ContainsFunc(dst.Pix, func(v uint8) bool { return v != 0 }) {
			t.Errorf("DecodeDraw at punch %v drew pixels", punch)
		}
	}
}

// TestDecodeMatchesDraw requires Decode, Decoder.Decode, DecodeDraw and
// Decoder.DecodeDraw to draw what codec.Draw draws, at a size that is not
// square and a punch other than 1.
func TestDecodeMatchesDraw(t *testing.T) {
	want := image.NewNRGBA(image.Rect(0, 0, 7, 5))
	_ = codec.Draw(want, hash4x3, 1.5)

	var d Decoder

	byDecode, _ := Decode(hash4x3, 7, 5, 1.5)
	byDecoder, _ := d.Decode(hash4x3, 7, 5, 1.5)
	byDraw := image.NewNRGBA(want.Rect)
	_ = DecodeDraw(byDraw, hash4x3, 1.5)
	byDecoderDraw := image.NewNRGBA(want.Rect)
	_ = d.DecodeDraw(byDecoderDraw, hash4x3, 1.5)

	for i, got := range []*image.NRGBA{byDecode, byDecoder, byDraw, byDecoderDraw} {
		if got == nil || got.Rect != want.Rect || !bytes.Equal(got.Pix, want.Pix) {
			t.Errorf("entry point %d draws differently from codec.Draw", i)
		}
	}
}

func TestDecodeDrawAllocs(t *testing.T) {
	dst := image.NewNRGBA(image.Rect(0, 0, 32, 32))

	var d Decoder

	if n := testing.AllocsPerRun(100, func() { _ = d.DecodeDraw(dst, hash4x3, 1) }); n != 0 {
		t.Errorf("Decoder.DecodeDraw allocates %v times, want 0", n)
	}

	if raceEnabled {
		return
	}

	if n := testing.AllocsPerRun(100, func() { _ = DecodeDraw(dst, hash4x3, 1) }); n != 0 {
		t.Errorf("DecodeDraw allocates %v times, want 0", n)
	}
}

// FuzzDecode requires Decode to accept only sizes from 1 to MaxSize, and
// Decode and DecodeDraw to agree on which inputs they accept and on the pixels.
func FuzzDecode(f *testing.F) {
	for _, s := range []struct {
		hash          string
		width, height int
		punch         float64
	}{
		{hash4x3, 7, 5, 1}, {hash1x1, 1, 1, 0}, {"00~~~~", 4, 4, 1}, {hash4x3, 0, 5, 1},
		{hash1x1, MaxSize + 1, 1, 1}, {hash4x3, 8, 8, math.NaN()}, {hash4x3, 8, 8, 1e308},
	} {
		f.Add(s.hash, s.width, s.height, s.punch)
	}

	f.Fuzz(func(t *testing.T, hash string, width, height int, punch float64) {
		sized := width > 0 && height > 0 && width <= MaxSize && height <= MaxSize
		if sized && width*height > 1<<16 {
			return // the large accepted sizes are TestDecodeSize's rows
		}

		img, err := Decode(hash, width, height, punch)
		if err == nil && !sized {
			t.Fatalf("%q at %dx%d: accepted", hash, width, height)
		}

		if !sized {
			return
		}

		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		if errDraw := DecodeDraw(dst, hash, punch); (errDraw == nil) != (err == nil) {
			t.Fatalf("%q at %dx%d, punch %v: Decode %v, DecodeDraw %v", hash, width, height, punch, err, errDraw)
		}

		if err == nil && !bytes.Equal(dst.Pix, img.Pix) {
			t.Fatalf("%q at %dx%d, punch %v: Decode and DecodeDraw draw differently", hash, width, height, punch)
		}
	})
}

// wide is a destination with any bounds, which draws nothing.
type wide struct{ image.Rectangle }

func (wide) Set(int, int, color.Color) {}
