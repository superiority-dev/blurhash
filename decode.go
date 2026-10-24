package blurhash

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// A Decoder decodes BlurHashes and reuses its memory between calls. The zero
// value is ready to use. A Decoder should be used by one goroutine at a time;
// [Decode] and [DecodeDraw] are safe for concurrent use.
type Decoder struct {
	d codec.Decoder
}

// NewDecoder returns a Decoder.
func NewDecoder() *Decoder {
	return new(Decoder)
}

// Decode returns hash drawn as a width by height image, as [Decode] does.
func (d *Decoder) Decode(hash string, width, height int, punch float64) (*image.NRGBA, error) {
	img, err := newImage(width, height)
	if err != nil {
		return nil, err
	}

	if err := d.d.Draw(img, hash, punch); err != nil {
		return nil, err
	}

	return img, nil
}

// DecodeDraw decodes hash into dst, as [DecodeDraw] does.
func (d *Decoder) DecodeDraw(dst draw.Image, hash string, punch float64) error {
	return d.d.Draw(dst, hash, punch)
}

// Components returns the numbers of horizontal and vertical components
// hash encodes, or [ErrInvalidHash].
func Components(hash string) (x, y int, err error) {
	return codec.Components(hash)
}

// Decode returns hash drawn as a width by height image, or
// [ErrInvalidDimensions], [ErrInvalidHash] or [ErrInvalidPunch]. It is safe
// for concurrent use.
//
// Punch scales the contrast: 1 draws the hash as encoded, larger values
// strengthen it, smaller ones soften it, 0 leaves only the average color
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
// [Decode], or returns [ErrInvalidHash] or [ErrInvalidPunch]. It is safe for
// concurrent use.
func DecodeDraw(dst draw.Image, hash string, punch float64) error {
	return codec.Draw(dst, hash, punch)
}

func newImage(width, height int) (*image.NRGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("%w: got %dx%d", ErrInvalidDimensions, width, height)
	}

	return image.NewNRGBA(image.Rect(0, 0, width, height)), nil
}
