package testimage

import (
	"image"
	"math/rand/v2"
)

// Random returns an opaque w by h image whose channels each take one of
// levels evenly spaced values.
func Random(seed uint64, w, h, levels int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	r := rand.New(rand.NewPCG(seed, 0))

	for i := range img.Pix {
		img.Pix[i] = uint8(r.IntN(levels) * 255 / max(1, levels-1))

		if i%4 == 3 {
			img.Pix[i] = 255
		}
	}

	return img
}
