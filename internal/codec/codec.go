// Package codec computes and draws BlurHashes for package blurhash.
package codec

import "errors"

var (
	// ErrInvalidHash is returned for a string that is not a BlurHash.
	ErrInvalidHash = errors.New("blurhash: invalid hash")

	// ErrInvalidPunch is returned for a punch that is NaN or infinite.
	ErrInvalidPunch = errors.New("blurhash: punch must be finite")
)
