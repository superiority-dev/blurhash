package blurhash

import (
	"fmt"
	"image"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// An Encoder computes BlurHashes and reuses its memory between calls, so that
// encoding a stream of images allocates only the hashes. The zero value is
// ready to use. An Encoder should be used by one goroutine at a time; [Encode]
// is safe for concurrent use.
type Encoder struct {
	// MaxSize is the largest width or height Encode accepts; zero or less
	// means the package's [MaxSize]. Encode's memory grows with the width and
	// height, not with their product.
	MaxSize int

	e codec.Encoder
}

// NewEncoder returns an Encoder.
func NewEncoder() *Encoder {
	return new(Encoder)
}

// Encode returns the BlurHash of img, as [Encode] does.
func (e *Encoder) Encode(xComponents, yComponents int, img image.Image) (string, error) {
	if err := check(xComponents, yComponents, img, sizeLimit(e.MaxSize)); err != nil {
		return "", err
	}

	return e.e.Encode(xComponents, yComponents, img), nil
}

// Encode returns the BlurHash of img using xComponents horizontal and
// yComponents vertical components, or [ErrInvalidComponents],
// [ErrInvalidDimensions] or [ErrImageTooLarge]. It is safe for concurrent
// use.
//
// Color is the premultiplied value [image/color.Color.RGBA] reports, cut to
// 8 bits; alpha is otherwise ignored.
func Encode(xComponents, yComponents int, img image.Image) (string, error) {
	if err := check(xComponents, yComponents, img, MaxSize); err != nil {
		return "", err
	}

	return codec.Encode(xComponents, yComponents, img), nil
}

func check(xComponents, yComponents int, img image.Image, limit int) error {
	if xComponents < codec.MinComponents || xComponents > codec.MaxComponents ||
		yComponents < codec.MinComponents || yComponents > codec.MaxComponents {
		return fmt.Errorf(gotSize, ErrInvalidComponents, xComponents, yComponents)
	}

	b := img.Bounds()
	if b.Empty() {
		return fmt.Errorf(gotSize, ErrInvalidDimensions, b.Dx(), b.Dy())
	}

	return checkSize(b, limit)
}
