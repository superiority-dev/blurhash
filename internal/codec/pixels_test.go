package codec

import (
	"fmt"
	"image"
	"image/color"
	"math/rand/v2"
	"testing"
)

// TestPremultiply requires premultiply to match color.NRGBA's RGBA method at
// every value and alpha.
func TestPremultiply(t *testing.T) {
	for a := range 256 {
		for v := range 256 {
			r, _, _, _ := color.NRGBA{uint8(v), 0, 0, uint8(a)}.RGBA()
			if got, want := premultiply(uint8(v), uint32(a)), uint8(r>>8); got != want {
				t.Fatalf("premultiply(%d, %d) = %d, color.NRGBA gives %d", v, a, got, want)
			}
		}
	}
}

// TestLinearRow requires linearRow's YCbCr read to equal At's, and the
// coefficients to match slowCoefficients', at every subsampling ratio,
// including sub-images whose origin is not a multiple of the chroma block.
func TestLinearRow(t *testing.T) {
	ratios := []image.YCbCrSubsampleRatio{
		image.YCbCrSubsampleRatio444, image.YCbCrSubsampleRatio422, image.YCbCrSubsampleRatio420,
		image.YCbCrSubsampleRatio440, image.YCbCrSubsampleRatio411, image.YCbCrSubsampleRatio410,
	}

	r := rand.New(rand.NewPCG(11, 0))

	for _, ratio := range ratios {
		img := image.NewYCbCr(image.Rect(-3, 2, 37, 29), ratio)
		for _, p := range [][]uint8{img.Y, img.Cb, img.Cr} {
			for i := range p {
				p[i] = uint8(r.IntN(256))
			}
		}

		// The whole image, a sub-image at an odd origin, and a single pixel.
		for _, rect := range []image.Rectangle{
			img.Rect, image.Rect(-1, 3, 30, 28), image.Rect(5, 7, 6, 8),
		} {
			sub := img.SubImage(rect)
			id := fmt.Sprintf("%v, %v", ratio, rect)

			var e Encoder

			for _, comp := range [][2]int{{4, 3}, {1, 1}, {9, 9}} {
				checkCoefficients(t, &e, id, comp[0], comp[1], sub)
			}

			// The rows read directly equal At's.
			line := make([][3]float64, rect.Dx())
			for y := range rect.Dy() {
				linearRow(line, sub, y)

				for x := range line {
					cr, cg, cb, _ := sub.At(rect.Min.X+x, rect.Min.Y+y).RGBA()

					want := [3]float64{linear[cr>>8], linear[cg>>8], linear[cb>>8]}
					if line[x] != want {
						t.Fatalf("%s: pixel (%d, %d) reads %v, At gives %v", id, x, y, line[x], want)
					}
				}
			}
		}
	}
}
