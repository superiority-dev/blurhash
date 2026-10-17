//go:build race

package blurhash

// raceEnabled reports a -race build, in which sync.Pool drops some of what
// it is given.
const raceEnabled = true
