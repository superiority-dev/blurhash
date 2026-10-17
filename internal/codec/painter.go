package codec

import (
	"image"
	"image/color"
	"image/draw"
)

// painter writes decoded pixels into one destination.
type painter struct {
	dst    draw.Image
	b      image.Rectangle
	pix    []uint8 // the NRGBA or RGBA bytes of b, or nil
	stride int
}

// drawRow draws pixel row y from its horizontal component values r.
// Columns x and w-x share products the same way rows do.
func (p *painter) drawRow(cosX [][]float64, r [][3]float64, y int) {
	w := p.b.Dx()

	for x := 0; 2*x <= w; x++ {
		var er, eg, eb, or, og, ob float64

		for i := 0; i < len(r); i += 2 {
			cx := cosX[i][x]
			er += float64(cx * r[i][0])
			eg += float64(cx * r[i][1])
			eb += float64(cx * r[i][2])
		}

		for i := 1; i < len(r); i += 2 {
			cx := cosX[i][x]
			or += float64(cx * r[i][0])
			og += float64(cx * r[i][1])
			ob += float64(cx * r[i][2])
		}

		p.set(x, y, er+or, eg+og, eb+ob)

		if x > 0 && 2*x != w {
			p.set(w-x, y, er-or, eg-og, eb-ob)
		}
	}
}

// set writes the pixel at (x, y) from its linear values.
func (p *painter) set(x, y int, r, g, b float64) {
	qr, qg, qb := toSRGB(r), toSRGB(g), toSRGB(b)

	if p.pix != nil {
		o := y*p.stride + 4*x
		p.pix[o], p.pix[o+1], p.pix[o+2], p.pix[o+3] = uint8(qr), uint8(qg), uint8(qb), 255

		return
	}

	p.dst.Set(p.b.Min.X+x, p.b.Min.Y+y, color.NRGBA{uint8(qr), uint8(qg), uint8(qb), 255})
}
