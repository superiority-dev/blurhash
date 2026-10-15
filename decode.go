package blurhash

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// Components returns the numbers of horizontal and vertical components
// hash encodes.
func Components(hash string) (x, y int, err error) {
	return codec.Components(hash)
}

// Decode returns hash drawn as a width by height image. It is safe for
// concurrent use.
//
// Punch scales the contrast: 1 draws the hash as encoded, larger values
// strengthen it, smaller ones soften it, 0 leaves only the average colour
// and a negative punch inverts it. Other decoders may adjust punch before
// applying it, so the same value can draw differently.
func Decode(hash string, width, height int, punch float64) (*image.NRGBA, error) {
	img, err := newImage(width, height)
	if err != nil {
		return nil, err
	}

	if err := codec.Draw(img, hash, punch); err != nil {
		return nil, err
	}

	return img, nil
}

// DecodeDraw decodes hash into dst, filling its bounds, with punch as for
// [Decode]. It is safe for concurrent use.
func DecodeDraw(dst draw.Image, hash string, punch float64) error {
	return codec.Draw(dst, hash, punch)
}

func newImage(width, height int) (*image.NRGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("%w: got %dx%d", ErrInvalidDimensions, width, height)
	}

	return image.NewNRGBA(image.Rect(0, 0, width, height)), nil
}
