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

func check(xComponents, yComponents int, img image.Image) error {
	if xComponents < 1 || xComponents > 9 || yComponents < 1 || yComponents > 9 {
		return fmt.Errorf("%w: got %dx%d", ErrInvalidComponents, xComponents, yComponents)
	}

	if b := img.Bounds(); b.Empty() {
		return fmt.Errorf("%w: got %dx%d", ErrInvalidDimensions, b.Dx(), b.Dy())
	}

	return nil
}
