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
