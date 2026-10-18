# blurhash

Fast BlurHash for Go, with the same result on every machine.

A [BlurHash](https://blurha.sh/) is a short string like
`L9FastAndExactBlurHashForGo.` that decodes into a blurred preview of an image,
shown while the real one loads. This package encodes and decodes them in pure Go,
with no dependencies.

![An image that this package encodes to L9FastAndExactBlurHashForGo.](.github/example.png) ![L9FastAndExactBlurHashForGo. decoded](.github/example-blurred.png)

```
go get github.com/superiority-dev/blurhash
```

## Why this package

- **Fast.** Hashing a photo is at least 2.8 times faster than with the other Go
  packages, and 48 to 407 times faster than with the original BlurHash C code.
- **The same hash everywhere.** An image gives the same hash on ARM as on Intel
  or AMD.
- **Accurate.** Matches the original C code's hashes almost everywhere, and is
  never less accurate where it doesn't.

## Speed

Time on real photos, as a multiple of this package's time.

**Encoding**

| package | language | ARM | Intel or AMD |
|---|---|---|---|
| [bbrks/go-blurhash](https://github.com/bbrks/go-blurhash) | Go | 2.8x to 13x | 3.1x to 17x |
| [buckket/go-blurhash](https://github.com/buckket/go-blurhash) | Go | 8.5x to 28x | 7.7x to 34x |
| [woltapp/blurhash](https://github.com/woltapp/blurhash) | C | 77x to 407x | 48x to 345x |
| [blurhash-python](https://pypi.org/project/blurhash-python/) | Python | 116x to 471x | 87x to 460x |
| [woltapp/blurhash](https://www.npmjs.com/package/blurhash) | TypeScript | 383x to 2,023x | 338x to 2,349x |
| [halcy/blurhash-python](https://pypi.org/project/blurhash/) | Python | 894x to 4,286x | 868x to 5,524x |

**Decoding**

| package | language | ARM | Intel or AMD |
|---|---|---|---|
| [bbrks/go-blurhash](https://github.com/bbrks/go-blurhash) | Go | 1.45x to 6.1x | 1.9x to 8.0x |
| [buckket/go-blurhash](https://github.com/buckket/go-blurhash) | Go | 24x to 84x | 31x to 109x |

The gap grows with the number of components.

How it was measured, the exact versions and every number are in
[BENCH.md](BENCH.md).

## Usage

### Encode

```go
hash, err := blurhash.Encode(4, 3, img) // 4x3 components: colour bands across and down, 1 to 9 each

var e blurhash.Encoder // reuses its memory; one goroutine at a time
for _, img := range images {
	hash, err := e.Encode(4, 3, img)
	// ...
}
```

Fastest with `*image.RGBA`, `*image.NRGBA` and `*image.YCbCr`, which
`image/jpeg` returns. Other image types work too, but slower.

### Decode

```go
img, err := blurhash.Decode(hash, 32, 24, 1) // a new image on every call

var wg sync.WaitGroup
for _, batch := range batches {
	wg.Go(func() {
		dst := image.NewNRGBA(image.Rect(0, 0, 32, 24)) // reused: DecodeDraw allocates nothing
		for _, hash := range batch {
			if err := blurhash.DecodeDraw(dst, hash, 1); err != nil {
				log.Print(err)
				continue
			}
			use(dst) // finish with dst, or copy it, before the next decode
		}
	})
}
wg.Wait()
```

## Shrink the image first

Encode a small copy, not the original: BlurHash keeps only the broad colours
anyway. This package does not resize. With
[golang.org/x/image/draw](https://pkg.go.dev/golang.org/x/image/draw):

```go
// Fit inside 100x100, keeping the aspect ratio.
b := img.Bounds()
w, h := 100, max(1, 100*b.Dy()/b.Dx())
if b.Dy() > b.Dx() {
	w, h = max(1, 100*b.Dx()/b.Dy()), 100
}

small := image.NewRGBA(image.Rect(0, 0, w, h))
draw.ApproxBiLinear.Scale(small, small.Bounds(), img, b, draw.Src, nil)

hash, err := blurhash.Encode(4, 3, small)
```

**Why 100 pixels.** The best balance between speed and quality: almost the same
result as the full image, in about 35 µs.

**Why ApproxBiLinear.** It comes closest to the full image of any x/image/draw
scaler. BiLinear and CatmullRom make it darker.

**Matching hashes made elsewhere.** Resize the same way everywhere. A different
scaler, or a size 1 pixel off, can give a different hash.

## Differences from the original

9,508 of 9,739 test hashes match the original C code exactly. In the other 231,
each value that differs is off by 1 step. Almost all of them sit exactly halfway
between 2 steps, where both answers are equally right. In the rest, this package
is the exact one. Every difference is listed in
[testdata/differences.txt](testdata/differences.txt).
