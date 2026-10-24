// ====================================================
// Copyright (C) 1993 by Sun Microsystems, Inc. All rights reserved.
//
// Developed at SunPro, a Sun Microsystems, Inc. business.
// Permission to use, copy, modify, and distribute this
// software is freely granted, provided that this notice
// is preserved.
// ====================================================
//
// The algorithm and constants below are from fdlibm's s_cos.c, e_rem_pio2.c,
// k_cos.c and k_sin.c, as V8 carries them in src/base/ieee754.cc at tag
// 12.4.254.21, commit 8ab91836885699b9ef686c3a30df4778501ef7d1.

// Package fdlibm holds fdlibm's cosine, which gives the same bits on every
// build target because no product is left for the compiler to fuse.
package fdlibm

import "math"

// The constants of __ieee754_rem_pio2, __kernel_cos and __kernel_sin.
var (
	invpio2 = math.Float64frombits(0x3fe45f306dc9c883) // 53 bits of 2/pi
	pio21   = math.Float64frombits(0x3ff921fb54400000) // first 33 bits of pi/2
	pio21t  = math.Float64frombits(0x3dd0b4611a626331) // pi/2 - pio21
	pio22   = math.Float64frombits(0x3dd0b4611a600000) // second 33 bits of pi/2
	pio22t  = math.Float64frombits(0x3ba3198a2e037073) // pi/2 - (pio21+pio22)
	pio23   = math.Float64frombits(0x3ba3198a2e000000) // third 33 bits of pi/2
	pio23t  = math.Float64frombits(0x397b839a252049c1) // pi/2 - (pio21+pio22+pio23)

	c1 = math.Float64frombits(0x3fa555555555554c)
	c2 = math.Float64frombits(0xbf56c16c16c15177)
	c3 = math.Float64frombits(0x3efa01a019cb1590)
	c4 = math.Float64frombits(0xbe927e4f809c52ad)
	c5 = math.Float64frombits(0x3e21ee9ebdb4b1c4)
	c6 = math.Float64frombits(0xbda8fae9be8838d4)

	s1 = math.Float64frombits(0xbfc5555555555549)
	s2 = math.Float64frombits(0x3f8111111110f8a6)
	s3 = math.Float64frombits(0xbf2a01a019c161d5)
	s4 = math.Float64frombits(0x3ec71de357b1fe7d)
	s5 = math.Float64frombits(0xbe5ae5e68a2b9ceb)
	s6 = math.Float64frombits(0x3de5d93a5acfd57c)
)

// npio2hw holds the high words of n*pi/2 for n from 1 to 32.
var npio2hw = [32]int32{
	0x3ff921fb, 0x400921fb, 0x4012d97c, 0x401921fb, 0x401f6a7a, 0x4022d97c,
	0x4025fdbb, 0x402921fb, 0x402c463a, 0x402f6a7a, 0x4031475c, 0x4032d97c,
	0x40346b9c, 0x4035fdbb, 0x40378fdb, 0x403921fb, 0x403ab41b, 0x403c463a,
	0x403dd85a, 0x403f6a7a, 0x40407e4c, 0x4041475c, 0x4042106c, 0x4042d97c,
	0x4043a28c, 0x40446b9c, 0x404534ac, 0x4045fdbb, 0x4046c6cb, 0x40478fdb,
	0x404858eb, 0x404921fb,
}

// Cos returns the cosine of x, for x from 0 to 2^19*pi/2, as fdlibm's cos
// computes it. Every product that feeds an addition or a subtraction is
// rounded before it is used, so no target fuses the two.
func Cos(x float64) float64 {
	if highWord(x)&0x7fffffff <= 0x3fe921fb { // |x| ~<= pi/4
		return kernelCos(x, 0)
	}

	n, y0, y1 := remPio2(x)

	switch n & 3 {
	case 0:
		return kernelCos(y0, y1)
	case 1:
		return -kernelSin(y0, y1)
	case 2:
		return -kernelCos(y0, y1)
	default:
		return kernelSin(y0, y1)
	}
}

// remPio2 is __ieee754_rem_pio2 for pi/4 < x <= 2^19*pi/2. It returns n and
// the remainder of x - n*pi/2 as y0 + y1.
func remPio2(x float64) (int32, float64, float64) {
	ix := highWord(x) & 0x7fffffff

	if ix < 0x4002d97c { // |x| < 3pi/4, special case with n = 1
		z := x - pio21
		if ix != 0x3ff921fb { // 33+53 bit pi is good enough
			y0 := z - pio21t
			return 1, y0, (z - y0) - pio21t
		}

		// Near pi/2, use 33+33+53 bit pi.
		z -= pio22
		y0 := z - pio22t

		return 1, y0, (z - y0) - pio22t
	}

	n := int32(float64(x*invpio2) + 0.5)
	fn := float64(n)
	r := x - float64(fn*pio21)
	w := float64(fn * pio21t) // 1st round good to 85 bit
	y0 := r - w

	if n >= 32 || ix == npio2hw[n-1] {
		j := ix >> 20

		if j-exponent(y0) > 16 { // 2nd iteration needed, good to 118
			t := r
			w = float64(fn * pio22)
			r = t - w
			w = float64(fn*pio22t) - ((t - r) - w)
			y0 = r - w

			if j-exponent(y0) > 49 { // 3rd iteration needed, 151 bits acc
				t = r
				w = float64(fn * pio23)
				r = t - w
				w = float64(fn*pio23t) - ((t - r) - w)
				y0 = r - w
			}
		}
	}

	return n, y0, (r - y0) - w
}

// kernelCos is __kernel_cos: cos(x + y) for |x| ~<= pi/4, y a tail of x.
func kernelCos(x, y float64) float64 {
	ix := highWord(x) & 0x7fffffff
	if ix < 0x3e400000 && int(x) == 0 { // |x| < 2^-27
		return 1
	}

	z := x * x
	r := z * (c1 + float64(z*(c2+float64(z*(c3+float64(z*(c4+float64(z*(c5+float64(z*c6))))))))))

	if ix < 0x3fd33333 { // |x| < 0.3
		return 1 - (float64(0.5*z) - (float64(z*r) - float64(x*y)))
	}

	qx := 0.28125
	if ix <= 0x3fe90000 { // x <= 0.78125
		qx = math.Float64frombits(uint64(uint32(ix-0x00200000)) << 32) // x/4
	}

	iz := float64(0.5*z) - qx
	a := 1 - qx

	return a - (iz - (float64(z*r) - float64(x*y)))
}

// kernelSin is __kernel_sin with iy = 1, the only form cos calls: sin(x + y)
// for |x| ~<= pi/4, y a tail of x.
func kernelSin(x, y float64) float64 {
	if highWord(x)&0x7fffffff < 0x3e400000 && int(x) == 0 { // |x| < 2^-27
		return x
	}

	z := x * x
	v := z * x
	r := s2 + float64(z*(s3+float64(z*(s4+float64(z*(s5+float64(z*s6)))))))

	return x - ((float64(z*(float64(0.5*y)-float64(v*r))) - y) - float64(v*s1))
}

func highWord(x float64) int32 {
	return int32(math.Float64bits(x) >> 32)
}

// exponent returns the biased exponent field of x.
func exponent(x float64) int32 {
	return int32(math.Float64bits(x) >> 52 & 0x7ff)
}
