package blurhash

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"image"
	"testing"

	"github.com/superiority-dev/blurhash/internal/testimage"
)

var update = flag.Bool("update", false, "rewrite the files in testdata from the current encoder")

// testdata is embedded so the tests run where only the test binary is.
//
//go:embed testdata
var testdata embed.FS

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

// Every hash that differs from comparison.txt must be listed in
// testdata/differences.txt, so a new difference and a vanished one both fail.
func TestEncodeComparison(t *testing.T) {
	file, err := testimage.Comparisons()
	if err != nil {
		t.Fatal(err)
	}

	var got []string

	n := 0

	for _, c := range testimage.Cases() {
		img := c.Image()

		for _, comp := range c.Components {
			h, err := Encode(comp.X, comp.Y, img)
			if err != nil {
				t.Fatalf("%s at %v: %v", c.ID, comp, err)
			}

			if h != file.Results[n].Hash {
				got = append(got, fmt.Sprintf("%s %v %s", c.ID, comp, h))
			}

			n++
		}
	}

	testimage.Golden(t, testdata, "testdata/differences.txt", got, *update)
}
