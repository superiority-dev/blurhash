// Package blurhash encodes images as BlurHash strings and decodes them into
// blurred placeholders. The format is defined at https://github.com/woltapp/blurhash.
package blurhash

import (
	"errors"
	"fmt"
	"image"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// MaxSize is the largest width or height Encode, Decode and DecodeDraw accept,
// and the default for [Encoder.MaxSize] and [Decoder.MaxSize].
const MaxSize = 4096

var (
	// ErrImageTooLarge is returned for an image wider or taller than
	// [MaxSize], or than an Encoder's or Decoder's MaxSize.
	ErrImageTooLarge = errors.New("blurhash: image too large")

	// ErrInvalidComponents is returned when a component count is outside
	// [1, 9].
	ErrInvalidComponents = errors.New("blurhash: must have between 1 and 9 components")

	// ErrInvalidDimensions is returned for an image with no pixels.
	ErrInvalidDimensions = errors.New("blurhash: width and height must be positive")

	// ErrInvalidHash is returned for a string that is not a BlurHash.
	ErrInvalidHash = codec.ErrInvalidHash

	// ErrInvalidPunch is returned when the punch, Decode's contrast setting, is
	// NaN or infinite.
	ErrInvalidPunch = codec.ErrInvalidPunch
)

// checkSize returns ErrImageTooLarge for a b that is not empty and is wider
// or taller than limit. The sizes are compared as uint, so a side that does
// not fit in an int is too large too.
func checkSize(b image.Rectangle, limit int) error {
	if s := b.Size(); !b.Empty() && (uint(s.X) > uint(limit) || uint(s.Y) > uint(limit)) {
		return fmt.Errorf("%w: %dx%d, need at most %dx%d", ErrImageTooLarge, uint(s.X), uint(s.Y), limit, limit)
	}

	return nil
}

// sizeLimit returns an Encoder's or Decoder's MaxSize field, or the
// package's MaxSize when the field is zero or less.
func sizeLimit(field int) int {
	if field <= 0 {
		return MaxSize
	}

	return field
}
