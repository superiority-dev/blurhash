package fdlibm

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"maps"
	"math"
	"math/big"
	"slices"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

// domainDigest is the SHA-256 of the big-endian bits of Cos at every
// argument the encoder uses for dimensions up to 1024: cos(pi*k*p/n) for k
// below 9 and p below n. Every build target must reproduce it.
const domainDigest = "a5459843dfe66a0d7ea00dd93cbdad550f03581df8a5c39ad9383520a50ee29a"

const prec = 256

// TestCosDomain requires Cos to return domainDigest's bits on every build
// target, at every argument the encoder uses for sides up to 1024.
func TestCosDomain(t *testing.T) {
	h := sha256.New()

	for n := 1; n <= 1024; n++ {
		for k := range 9 {
			for p := range n {
				bits := math.Float64bits(Cos(math.Pi * float64(k) * float64(p) / float64(n)))
				_, _ = h.Write(binary.BigEndian.AppendUint64(nil, bits))
			}
		}
	}

	if got := hex.EncodeToString(h.Sum(nil)); got != domainDigest {
		t.Errorf("cosines have digest %s, want %s", got, domainDigest)
	}
}

// TestCosAccuracy requires Cos to be within one ulp of the exact cosine at
// every argument the corpus images produce.
func TestCosAccuracy(t *testing.T) {
	sizes := map[int]bool{}

	for _, c := range testimage.Cases() {
		b := c.Image().Bounds()
		sizes[b.Dx()] = true
		sizes[b.Dy()] = true
	}

	for _, n := range slices.Sorted(maps.Keys(sizes)) {
		for k := range 9 {
			for p := range n {
				x := math.Pi * float64(k) * float64(p) / float64(n)
				got, want := Cos(x), exactCos(x)

				if got != want && math.Nextafter(got, want) != want {
					t.Errorf("Cos(pi*%d*%d/%d) = %x, exact rounds to %x", k, p, n, got, want)
				}
			}
		}
	}
}

// pi to 80 digits, more than prec needs.
var pi, _, _ = big.ParseFloat("3.14159265358979323846264338327950288419716939937510"+
	"582097494459230781640628620899", 10, prec, big.ToNearestEven)

// exactCos returns cos(x) for 0 <= x < 64, correctly rounded, by Taylor
// series after reduction into [0, 2pi).
func exactCos(x float64) float64 {
	bx := new(big.Float).SetPrec(prec).SetFloat64(x)
	twoPi := new(big.Float).SetPrec(prec).Mul(pi, big.NewFloat(2))

	for bx.Cmp(twoPi) >= 0 {
		bx.Sub(bx, twoPi)
	}

	x2 := new(big.Float).SetPrec(prec).Mul(bx, bx)
	sum := new(big.Float).SetPrec(prec).SetFloat64(1)
	term := new(big.Float).SetPrec(prec).SetFloat64(1)

	// Terms up to x^238/238!, far below 2^-256 for x below 2pi.
	for i := int64(1); i < 120; i++ {
		term.Mul(term, x2)
		term.Quo(term, new(big.Float).SetInt64((2*i-1)*(2*i)))
		term.Neg(term)
		sum.Add(sum, term)
	}

	f, _ := sum.Float64()

	return f
}
