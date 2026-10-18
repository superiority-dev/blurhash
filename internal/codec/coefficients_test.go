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

// coefficientDigest returns the SHA-256 of every coefficient from
// slowCoefficients and from an Encoder, at each of c's component counts.
func coefficientDigest(c testimage.Case) string {
	img := c.Image()
	h := sha256.New()

	for _, comp := range c.Components {
		slow := slowCoefficients(comp.X, comp.Y, img)
		fast := new(Encoder).coefficients(comp.X, comp.Y, img)

		for _, f := range [][][3]float64{slow, fast} {
			for _, rgb := range f {
				for _, v := range rgb {
					_, _ = h.Write(binary.BigEndian.AppendUint64(nil, math.Float64bits(v)))
				}
			}
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}
