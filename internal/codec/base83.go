package codec

import "strings"

const digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz#$%*+,-.:;=?@[]^_{|}~"

const base = len(digits)

// writeBase83 writes v as n base-83 digits, most significant first.
func writeBase83(b *strings.Builder, v, n int) {
	divisor := 1
	for range n - 1 {
		divisor *= base
	}

	for ; divisor > 0; divisor /= base {
		b.WriteByte(digits[v/divisor%base])
	}
}

// readBase83 returns the value of the base-83 digits in s, or false if s
// holds a character that is not one.
func readBase83(s string) (int, bool) {
	v := 0

	for i := range len(s) {
		d := strings.IndexByte(digits, s[i])
		if d < 0 {
			return 0, false
		}

		v = v*base + d
	}

	return v, true
}
