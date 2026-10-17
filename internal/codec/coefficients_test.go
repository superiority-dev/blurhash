package codec

import (
	"crypto/sha256"
	"embed"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"math"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

var update = flag.Bool("update", false, "rewrite the files in testdata from the current encoder")

// testdata is embedded so the tests run where only the test binary is.
//
//go:embed testdata
var testdata embed.FS

// TestCoefficientBits compares a digest of every coefficient's bits, per corpus
// image, with testdata/coefficients.txt. Hashes show an arithmetic change
// only where it crosses a quantisation boundary; this shows any change, and
// run on each build target it shows whether the targets agree.
func TestCoefficientBits(t *testing.T) {
	var lines []string

	for _, c := range testimage.Cases() {
		lines = append(lines, c.ID+" "+coefficientDigest(c))
	}

	testimage.Golden(t, testdata, "testdata/coefficients.txt", lines, *update)
}

// coefficientDigest returns the SHA-256 of the DC ties of c and, at each of
// its component counts, every coefficient from factors and from an Encoder,
// and the Encoder's bounds.
func coefficientDigest(c testimage.Case) string {
	img := c.Image()
	p := read(img)
	h := sha256.New()

	for _, d := range p.dcTie {
		_, _ = h.Write(binary.BigEndian.AppendUint64(nil, uint64(int64(d))))
	}

	write := func(v float64) { _, _ = h.Write(binary.BigEndian.AppendUint64(nil, math.Float64bits(v))) }

	for _, comp := range c.Components {
		for _, f := range factors(comp.X, comp.Y, p) {
			for _, v := range f {
				write(v)
			}
		}

		f, bound, _ := new(Encoder).coefficients(comp.X, comp.Y, img)
		for _, v := range f {
			for _, v := range v {
				write(v)
			}
		}

		for _, v := range bound {
			write(v)
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}
