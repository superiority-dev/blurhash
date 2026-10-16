package testimage

import "encoding/binary"

// splitMix is SplitMix64: the corpus must stay byte-stable across Go releases, which math/rand/v2 does not promise.
type splitMix struct {
	state uint64
	buf   [8]byte
	n     int
}

func newSplitMix(seed uint64) *splitMix {
	return &splitMix{state: seed}
}

func (r *splitMix) byte() uint8 {
	if r.n == 0 {
		r.state += 0x9e3779b97f4a7c15
		z := r.state
		z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
		z = (z ^ z>>27) * 0x94d049bb133111eb
		binary.LittleEndian.PutUint64(r.buf[:], z^z>>31)
		r.n = 8
	}

	r.n--

	return r.buf[r.n]
}
