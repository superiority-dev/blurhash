package blurhash

import (
	"bytes"
	"errors"
	"image"
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
