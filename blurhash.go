// Package blurhash encodes images as BlurHash strings and decodes them into
// blurred placeholders. The format is defined at https://github.com/woltapp/blurhash.
package blurhash

import (
	"errors"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// MaxSize is the largest width or height Encode, Decode and DecodeDraw accept.
const MaxSize = 4096

var (
	// ErrImageTooLarge is returned for an image wider or taller than
	// [MaxSize].
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
