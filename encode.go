package blurhash

import (
	"fmt"
	"image"

	"github.com/superiority-dev/blurhash/internal/codec"
)

// Encode returns the BlurHash of img using xComponents horizontal and
// yComponents vertical components, each between 1 and 9. It is safe for
// concurrent use.
//
// Colour is the premultiplied value [image/color.Color.RGBA] reports, cut to
// 8 bits; alpha is otherwise ignored.
func Encode(xComponents, yComponents int, img image.Image) (string, error) {
	if err := check(xComponents, yComponents, img); err != nil {
		return "", err
	}

	return codec.Encode(xComponents, yComponents, img), nil
}

// An Encoder computes BlurHashes and reuses its memory between calls, so that
// encoding a stream of images allocates only the hashes. The zero value is
// ready to use. An Encoder should be used by one goroutine at a time; [Encode]
// is safe for concurrent use.
type Encoder struct {
	e codec.Encoder
}

// NewEncoder returns an Encoder.
func NewEncoder() *Encoder {
	return new(Encoder)
}

// Encode returns the BlurHash of img, as [Encode] does.
func (e *Encoder) Encode(xComponents, yComponents int, img image.Image) (string, error) {
	if err := check(xComponents, yComponents, img); err != nil {
		return "", err
	}

	return e.e.Encode(xComponents, yComponents, img), nil
}

func check(xComponents, yComponents int, img image.Image) error {
	if xComponents < 1 || xComponents > 9 || yComponents < 1 || yComponents > 9 {
		return fmt.Errorf("%w: got %dx%d", ErrInvalidComponents, xComponents, yComponents)
	}

	if b := img.Bounds(); b.Empty() {
		return fmt.Errorf("%w: got %dx%d", ErrInvalidDimensions, b.Dx(), b.Dy())
	}

	return nil
}
