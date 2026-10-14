// Package blurhash encodes images as BlurHash strings. The format is defined
// at https://github.com/woltapp/blurhash.
package blurhash

import "errors"

var (
	// ErrInvalidComponents is returned when a component count is outside
	// [1, 9].
	ErrInvalidComponents = errors.New("blurhash: must have between 1 and 9 components")

	// ErrInvalidDimensions is returned for an image with no pixels.
	ErrInvalidDimensions = errors.New("blurhash: width and height must be positive")
)
