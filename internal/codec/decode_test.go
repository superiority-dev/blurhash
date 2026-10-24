package codec

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

// nrgbaBytes returns img's pixels as NRGBA bytes, row by row.
func nrgbaBytes(img image.Image) []uint8 {
	b := img.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Rect, img, b.Min, draw.Src)

	return out.Pix
}

// TestDecoderDraw requires Decoder.Draw to draw slowDecode's pixels byte for
// byte, for every fixture hash at several sizes and punches, into NRGBA, RGBA
// at an offset origin, and RGBA64.
func TestDecoderDraw(t *testing.T) {
	var d Decoder

	// Square, single-pixel, odd and thumbnail sizes; punches either side of 1,
	// zero, and negative.
	sizes := [][2]int{{32, 32}, {1, 1}, {7, 5}, {32, 24}, {2, 3}}
	punches := []float64{1, 0.5, 2, 0, -1}

	file, err := testimage.Comparisons()
	if err != nil {
		t.Fatal(err)
	}

	n := 0

	for k, hash := range file.Hashes() {
		s := sizes[k%len(sizes)]
		punch := punches[k/len(sizes)%len(punches)]

		dsts := []draw.Image{
			image.NewNRGBA(image.Rect(0, 0, s[0], s[1])),
			image.NewRGBA(image.Rect(3, -2, 3+s[0], -2+s[1])),
			image.NewRGBA64(image.Rect(0, 0, s[0], s[1])),
		}

		slow := image.NewNRGBA(image.Rect(0, 0, s[0], s[1]))
		if err := slowDecode(slow, hash, punch); err != nil {
			t.Fatal(err)
		}

		for _, dst := range dsts {
			if err := d.Draw(dst, hash, punch); err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(nrgbaBytes(dst), slow.Pix) {
				t.Fatalf("%s at %dx%d, punch %g, into %T: differs from slowDecode",
					hash, s[0], s[1], punch, dst)
			}

			n++
		}
	}

	t.Logf("%d decodes equal slowDecode's", n)
}

// TestDrawConcurrent requires Draw to draw slowDecode's images when called
// from several goroutines at once.
func TestDrawConcurrent(t *testing.T) {
	file, err := testimage.Comparisons()
	if err != nil {
		t.Fatal(err)
	}

	hashes := file.Hashes()[:2000]

	want := make([][]uint8, len(hashes))
	for i, h := range hashes {
		slow := image.NewNRGBA(image.Rect(0, 0, 12, 9))
		if err := slowDecode(slow, h, 1); err != nil {
			t.Fatal(err)
		}

		want[i] = slow.Pix
	}

	got := make([][][]uint8, 8)

	var wg sync.WaitGroup

	for g := range got {
		got[g] = make([][]uint8, len(hashes))

		wg.Go(func() {
			for _, i := range rand.New(rand.NewPCG(uint64(g), 1)).Perm(len(hashes)) {
				img := image.NewNRGBA(image.Rect(0, 0, 12, 9))
				if err := Draw(img, hashes[i], 1); err != nil {
					t.Error(err)

					return
				}

				got[g][i] = img.Pix
			}
		})
	}

	wg.Wait()

	for g := range got {
		for i := range hashes {
			if !slices.Equal(got[g][i], want[i]) {
				t.Errorf("goroutine %d, %s: differs from slowDecode", g, hashes[i])
			}
		}
	}
}

// FuzzDraw requires Draw to fail only with ErrInvalidHash or ErrInvalidPunch,
// to draw slowDecode's pixels except within decodeBound of a byte boundary,
// and every hash it draws to re-encode with the same components.
func FuzzDraw(f *testing.F) {
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

		slow := image.NewNRGBA(dst.Rect)
		_ = slowDecode(slow, hash, punch)

		// A difference is allowed only where the slow value is within
		// decodeBound of a byte boundary.
		var c [81][3]float64

		nx, ny, _ := readColors(&c, hash, punch)
		bound := decodeBound(c[:nx*ny])

		for i := range dst.Pix {
			if dst.Pix[i] == slow.Pix[i] {
				continue
			}

			x, y, ch := i/4%width, i/4/width, i%4
			v := slowPixel(c[:nx*ny], nx, width, height, x, y, ch)
			e := bound[ch]

			if q := toSRGB(v); toSRGB(v-e) == q && toSRGB(v+e) == q {
				t.Fatalf("%q, punch %v: (%d, %d) channel %d is %d, slow %d; %v is %g or more from a boundary",
					hash, punch, x, y, ch, dst.Pix[i], slow.Pix[i], v, e)
			}
		}

		x, y, _ := Components(hash)

		again := Encode(x, y, img)

		if ax, ay, err := Components(again); ax != x || ay != y || err != nil {
			t.Fatalf("%q re-encodes to %q, with components %d, %d, %v", hash, again, ax, ay, err)
		}
	})
}

// slowDecode draws hash into dst the plain way, which the Decoder is checked
// against: each pixel summed straight from the formula, component by
// component, rather than in two passes.
func slowDecode(dst draw.Image, hash string, punch float64) error {
	var c [81][3]float64

	nx, ny, err := readColors(&c, hash, punch)
	if err != nil {
		return err
	}

	b := dst.Bounds()
	w, h := b.Dx(), b.Dy()
	cosX := cosines(nx, w)
	cosY := cosines(ny, h)

	for y := range h {
		for x := range w {
			var sum [3]float64

			for j := range ny {
				for i := range nx {
					basis := cosX[i][x] * cosY[j][y]

					for ch, v := range c[j*nx+i] {
						// The product is rounded before it is added, so no target fuses the two.
						sum[ch] += float64(v * basis)
					}
				}
			}

			r, g, bl := toSRGB(sum[0]), toSRGB(sum[1]), toSRGB(sum[2])
			dst.Set(b.Min.X+x, b.Min.Y+y, color.NRGBA{uint8(r), uint8(g), uint8(bl), 255})
		}
	}

	return nil
}

// decodeBound is how far the Decoder's sums may lie from slowDecode's, per
// channel: 1e-12 times the channel's sum of absolute component values, about
// seven times the largest rounding gap the two orders allow at 9x9.
func decodeBound(c [][3]float64) [3]float64 {
	var out [3]float64

	for _, v := range c {
		for ch := range out {
			out[ch] += math.Abs(v[ch])
		}
	}

	for ch := range out {
		out[ch] *= 1e-12
	}

	return out
}

// TestDrawHugePunch requires Draw to draw every pixel, without a panic, at a
// punch large enough to overflow the sums.
func TestDrawHugePunch(t *testing.T) {
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

// slowPixel returns channel ch of the pixel at (x, y) of a w by h image,
// adding each component's color times its two cosines in index order.
func slowPixel(c [][3]float64, nx, w, h, x, y, ch int) float64 {
	var sum float64

	for k, v := range c {
		basis := cosine(k%nx, x, w) * cosine(k/nx, y, h)
		sum += float64(v[ch] * basis)
	}

	return sum
}
