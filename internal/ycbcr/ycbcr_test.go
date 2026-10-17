package ycbcr

import (
	"image/color"
	"testing"
)

// TestRGB holds RGB to color.YCbCr's RGBA method on every input.
func TestRGB(t *testing.T) {
	for i := range 1 << 24 {
		y, cb, cr := uint8(i>>16), uint8(i>>8), uint8(i)
		r, g, b := RGB(y, cb, cr)

		wr, wg, wb, _ := color.YCbCr{Y: y, Cb: cb, Cr: cr}.RGBA()
		if r != uint8(wr>>8) || g != uint8(wg>>8) || b != uint8(wb>>8) {
			t.Fatalf("RGB(%d, %d, %d) = %d, %d, %d; color.YCbCr gives %d, %d, %d",
				y, cb, cr, r, g, b, wr>>8, wg>>8, wb>>8)
		}
	}
}
