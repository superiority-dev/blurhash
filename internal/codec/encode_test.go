package codec

import "testing"

func TestDCTie(t *testing.T) {
	for _, c := range []struct {
		name    string
		s, w, n uint64
		want    int
	}{
		{"black, which is 0.5 and not a tie", 0, 0, 7, -1},
		{"3 and 4 split", 14, 0, 4, 4},
		{"5 white among 1933 black", 0, 5, 1938, 9},
		{"not a tie", 1, 0, 3, -1},
		{"white beyond the linear branch", 0, 1, 2, -1},
	} {
		if got := dcTie(c.s, c.w, c.n); got != c.want {
			t.Errorf("%s: dcTie(%d, %d, %d) = %d, want %d", c.name, c.s, c.w, c.n, got, c.want)
		}
	}
}
