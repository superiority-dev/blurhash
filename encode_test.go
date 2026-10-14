package blurhash

import (
	"errors"
	"image"
	"testing"
)

func TestEncodeErrors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))

	// Each argument just below 1 and just above 9, and a negative.
	for _, c := range []image.Point{{0, 1}, {1, 0}, {10, 1}, {1, 10}, {-1, 3}} {
		if _, err := Encode(c.X, c.Y, img); !errors.Is(err, ErrInvalidComponents) {
			t.Errorf("Encode(%d, %d): error %v, want ErrInvalidComponents", c.X, c.Y, err)
		}
	}

	// No pixels.
	for _, r := range []image.Rectangle{
		image.Rect(0, 0, 0, 0), // empty
		image.Rect(0, 0, 5, 0), // zero height
		image.Rect(3, 3, 3, 9), // zero width, away from the origin
	} {
		if _, err := Encode(4, 3, image.NewRGBA(r)); !errors.Is(err, ErrInvalidDimensions) {
			t.Errorf("Encode of %v: error %v, want ErrInvalidDimensions", r, err)
		}
	}
}
