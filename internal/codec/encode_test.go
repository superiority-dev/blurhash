package codec

import (
	"image"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

type sample struct {
	id     string
	img    image.Image
	nx, ny int
}

// samples returns every corpus encode and 2000 random ones.
func samples() []sample {
	var out []sample

	for _, c := range testimage.Cases() {
		img := c.Image()
		for _, comp := range c.Components {
			out = append(out, sample{c.ID, img, comp.X, comp.Y})
		}
	}

	r := rand.New(rand.NewPCG(0xb10b, 0))

	// Sizes up to 40x40, each byte drawn from 2, 3, 5 or 256 evenly spaced
	// levels: noise in a few values a channel, or over the full range.
	for range 2000 {
		w, h := 1+r.IntN(40), 1+r.IntN(40)
		img := testimage.Random(r.Uint64(), w, h, []int{2, 3, 5, 256}[r.IntN(4)])
		out = append(out, sample{"random", img, 1 + r.IntN(9), 1 + r.IntN(9)})
	}

	return out
}

// TestEncoderCoefficientsMatchesSlow requires one Encoder's coefficients to
// lie within tolerance of slowCoefficients' over every sample, its cache and
// buffers reused across sizes.
func TestEncoderCoefficientsMatchesSlow(t *testing.T) {
	var e Encoder

	for _, s := range samples() {
		checkCoefficients(t, &e, s.id, s.nx, s.ny, s.img)
	}
}

// TestEncodeConcurrent requires Encode to return a fresh Encoder's hashes when
// called from several goroutines at once.
func TestEncodeConcurrent(t *testing.T) {
	s := samples()

	want := make([]string, len(s))
	for i, s := range s {
		want[i] = new(Encoder).Encode(s.nx, s.ny, s.img)
	}

	got := make([][]string, 8)

	var wg sync.WaitGroup

	for g := range got {
		got[g] = make([]string, len(s))

		wg.Go(func() {
			for _, i := range rand.New(rand.NewPCG(uint64(g), 0)).Perm(len(s)) {
				got[g][i] = Encode(s[i].nx, s[i].ny, s[i].img)
			}
		})
	}

	wg.Wait()

	for g := range got {
		for i := range s {
			if got[g][i] != want[i] {
				t.Errorf("goroutine %d, %s at %dx%d: %s, a fresh Encoder %s",
					g, s[i].id, s[i].nx, s[i].ny, got[g][i], want[i])
			}
		}
	}
}

// fuzzEncoder is shared by every fuzz input a process runs, so its cache and
// buffers see a mixed stream.
var fuzzEncoder Encoder

// FuzzEncode requires an Encoder's coefficients to lie within tolerance of
// slowCoefficients', and Encode to return what Encoder.Encode returns, on
// random images and component counts.
func FuzzEncode(f *testing.F) {
	f.Add(uint64(1), uint8(8), uint8(8), uint8(4), uint8(3), uint8(2))
	f.Add(uint64(2), uint8(1), uint8(1), uint8(9), uint8(9), uint8(0))
	f.Add(uint64(3), uint8(33), uint8(7), uint8(1), uint8(9), uint8(255))

	f.Fuzz(func(t *testing.T, seed uint64, w, h, nx, ny, levels uint8) {
		img := testimage.Random(seed, 1+int(w%64), 1+int(h%64), 2+int(levels)%255)
		x, y := 1+int(nx%9), 1+int(ny%9)

		checkCoefficients(t, &fuzzEncoder, "random", x, y, img)

		if got, want := Encode(x, y, img), fuzzEncoder.Encode(x, y, img); got != want {
			t.Fatalf("Encode returns %s, Encoder.Encode %s", got, want)
		}
	})
}
