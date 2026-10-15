package codec

import (
	"errors"
	"image"
	"math"
	"testing"
)

func FuzzDecode(f *testing.F) {
	// The usual example, the smallest hash, the 9x9 size digit with too few
	// characters, and 2x1.
	for _, h := range []string{"LEHV6nWB2yk8pyo0adR*.7kCMdnj", "000000", "|~Lqe9", "10Lqe9fQ"} {
		f.Add(h, uint8(8), uint8(6), 1.0)
	}

	f.Fuzz(func(t *testing.T, hash string, w, h uint8, punch float64) {
		width, height := 1+int(w%40), 1+int(h%40)

		img := image.NewNRGBA(image.Rect(0, 0, width, height))
		if err := Draw(img, hash, 1); err != nil {
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("Draw(%q): error %v", hash, err)
			}

			return
		}

		dst := image.NewNRGBA(image.Rect(0, 0, width, height))
		if err := Draw(dst, hash, punch); err != nil {
			if !errors.Is(err, ErrInvalidPunch) {
				t.Fatalf("%q at punch %v: %v", hash, punch, err)
			}

			return
		}

		x, y, _ := Components(hash)

		again := Encode(x, y, img)

		if ax, ay, err := Components(again); ax != x || ay != y || err != nil {
			t.Fatalf("%q re-encodes to %q, with components %d, %d, %v", hash, again, ax, ay, err)
		}
	})
}

// TestDecodeHugePunch checks that a punch large enough to overflow the sums
// still draws every pixel, without a panic.
func TestDecodeHugePunch(t *testing.T) {
	for _, punch := range []float64{1e300, 1e306, -1e308, math.MaxFloat64} {
		img := image.NewNRGBA(image.Rect(0, 0, 9, 7))
		if err := Draw(img, "LEHV6nWB2yk8pyo0adR*.7kCMdnj", punch); err != nil {
			t.Fatal(err)
		}

		for i := 3; i < len(img.Pix); i += 4 {
			if img.Pix[i] != 255 {
				t.Fatalf("punch %g: pixel %d not drawn", punch, i/4)
			}
		}
	}
}
