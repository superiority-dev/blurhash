package codec

import (
	"image"
	"math"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

// tolerance is how far coefficients may lie from slowCoefficients: 1,000
// times the largest gap measured, 9.6e-13, on a 2000x1500 random image at
// 9x9.
const tolerance = 1e-9

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

// slowCoefficients returns the coefficients of img computed the plain way,
// which the Encoder is checked against: each pixel read through At rather
// than linearRow's direct reads, and each component summed straight from the
// formula rather than in two passes. Index j*nx+i is horizontal component i
// and vertical component j.
func slowCoefficients(nx, ny int, img image.Image) [][3]float64 {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	pix := make([][3]float64, w*h)

	for y := range h {
		for x := range w {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			pix[y*w+x] = [3]float64{linear[r>>8], linear[g>>8], linear[bl>>8]}
		}
	}

	cosX := cosines(nx, w)
	cosY := cosines(ny, h)

	out := make([][3]float64, nx*ny)

	for j := range ny {
		for i := range nx {
			var sum [3]float64

			for y := range h {
				for x := range w {
					basis := cosX[i][x] * cosY[j][y]

					for c, v := range pix[y*w+x] {
						// The product is rounded before it is added, so no target fuses the two.
						sum[c] += float64(basis * v)
					}
				}
			}

			// The average colour (i = j = 0) is the plain mean; every other
			// component is doubled, as the cosine transform's normalisation
			// requires. The scale is applied once, after the sum.
			normalisation := 2.0
			if i == 0 && j == 0 {
				normalisation = 1
			}

			scale := normalisation / float64(w*h)

			for c := range sum {
				out[j*nx+i][c] = sum[c] * scale
			}
		}
	}

	return out
}

// cosines returns cos(pi*k*p/n) for each component k below count and each
// position p below n.
func cosines(count, n int) [][]float64 {
	out := make([][]float64, count)

	for k := range out {
		out[k] = make([]float64, n)
		for p := range n {
			out[k][p] = cosine(k, p, n)
		}
	}

	return out
}

// TestEncoderCoefficients holds the Encoder's coefficients to
// slowCoefficients over the samples, with one Encoder throughout so its
// cache and buffers see every size.
func TestEncoderCoefficients(t *testing.T) {
	var e Encoder

	for _, s := range samples() {
		checkCoefficients(t, &e, s.id, s.nx, s.ny, s.img)
	}
}

// TestEncodeConcurrent runs Encode from 8 goroutines over the samples, keeps
// every hash and compares them with a fresh Encoder's after the last call, so
// a hash that shares a pooled buffer shows up even without the race detector.
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

// checkCoefficients fails t where a coefficient e computes for img lies more
// than tolerance from slowCoefficients'.
func checkCoefficients(t *testing.T, e *Encoder, id string, nx, ny int, img image.Image) {
	t.Helper()

	got := e.coefficients(nx, ny, img)

	for k, want := range slowCoefficients(nx, ny, img) {
		for c := range want {
			if d := math.Abs(got[k][c] - want[c]); d > tolerance {
				t.Fatalf("%s at %dx%d: coefficient %d channel %d is %g from slowCoefficients'",
					id, nx, ny, k, c, d)
			}
		}
	}
}
