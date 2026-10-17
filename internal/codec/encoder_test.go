package codec

import (
	"image"
	"math"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

// reference is Encode with the coefficients always taken from factors.
func reference(nx, ny int, img image.Image) string {
	p := read(img)
	hash, _ := format(nx, ny, factors(nx, ny, p), [3]float64{}, p.dcTie)

	return hash
}

type sample struct {
	id     string
	img    image.Image
	nx, ny int
}

// randomID is the id of every random sample.
const randomID = "random"

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
		out = append(out, sample{randomID, img, 1 + r.IntN(9), 1 + r.IntN(9)})
	}

	return out
}

func TestEncoderBound(t *testing.T) {
	worst := 0.0

	var e Encoder

	for _, s := range samples() {
		p := read(s.img)
		f, bound, dcTie := e.coefficients(s.nx, s.ny, s.img)

		if dcTie != p.dcTie {
			t.Fatalf("%s: DC ties %v, read gives %v", s.id, dcTie, p.dcTie)
		}

		for k, ref := range factors(s.nx, s.ny, p) {
			for c := range 3 {
				d := math.Abs(f[k][c] - ref[c])
				if d > bound[c] {
					t.Fatalf("%s at %dx%d: coefficient %d channel %d differs by %g, bound %g", s.id, s.nx, s.ny, k, c, d, bound[c])
				}

				if bound[c] > 0 {
					worst = max(worst, d/bound[c])
				}
			}
		}
	}

	t.Logf("largest difference, as a fraction of the bound: %.3g", worst)
}

// TestGuardMatters checks the guard is not idle: with every bound set to
// zero, some corpus hashes must differ from the reference.
func TestGuardMatters(t *testing.T) {
	var e Encoder

	differ := 0

	for _, s := range samples() {
		if s.id == randomID {
			continue
		}

		f, _, dcTie := e.coefficients(s.nx, s.ny, s.img)
		if hash, _ := format(s.nx, s.ny, f, [3]float64{}, dcTie); hash != reference(s.nx, s.ny, s.img) {
			differ++
		}
	}

	if differ == 0 {
		t.Fatal("without the guard every corpus hash still equals the reference")
	}

	t.Logf("without the guard, %d corpus hashes differ from the reference", differ)
}

func TestEncodeMatchesReference(t *testing.T) {
	var fallbacks, corpus, n int

	var e Encoder

	for _, s := range samples() {
		want := reference(s.nx, s.ny, s.img)

		if got := Encode(s.nx, s.ny, s.img); got != want {
			t.Errorf("%s at %dx%d: Encode returns %s; reference %s", s.id, s.nx, s.ny, got, want)
		}

		// One Encoder over the whole stream, so its cache sees every size.
		if got := e.Encode(s.nx, s.ny, s.img); got != want {
			t.Errorf("%s at %dx%d: Encoder.Encode returns %s; reference %s", s.id, s.nx, s.ny, got, want)
		}

		f, bound, dcTie := new(Encoder).coefficients(s.nx, s.ny, s.img)
		if _, ok := format(s.nx, s.ny, f, bound, dcTie); !ok {
			fallbacks++
			if s.id != randomID {
				corpus++
			}
		}

		n++
	}

	if fallbacks == corpus {
		t.Fatal("the random inputs produced no guard fallback, so the random share no longer covers the fallback path; pin a crafted fallback case rather than reseeding")
	}

	t.Logf("fell back to factors on %d of %d encodes, %d of them corpus encodes", fallbacks, n, corpus)
}

// TestEncodeConcurrent runs Encode from 8 goroutines over the samples, keeps
// every hash and compares them after the last call, so a hash that shares a
// pooled buffer shows up even without the race detector.
func TestEncodeConcurrent(t *testing.T) {
	s := samples()

	want := make([]string, len(s))
	for i, s := range s {
		want[i] = reference(s.nx, s.ny, s.img)
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
				t.Errorf("goroutine %d, %s at %dx%d: %s, reference %s", g, s[i].id, s[i].nx, s[i].ny, got[g][i], want[i])
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

		got := Encode(x, y, img)

		want := reference(x, y, img)
		if got != want {
			t.Fatalf("Encode returns %s, reference %s", got, want)
		}

		if got := fuzzEncoder.Encode(x, y, img); got != want {
			t.Fatalf("Encoder.Encode returns %s, reference %s", got, want)
		}
	})
}
