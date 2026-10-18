# Benchmarks

How the speed figures in the [README](README.md) were measured, and every number
behind them.

## What was timed

Encoding 6 real photos at 4 sizes. Go's `image/jpeg` decoded each photo once,
and every implementation got the same 8-bit RGB pixels. So no JPEG decoding is in
any number. This package was timed on its `*image.RGBA` path.

2 more tables cover only Go packages: encoding `*image.YCbCr`, which
`image/jpeg` returns, and decoding.

Each figure is an implementation's median time over 6 runs, as a multiple of
this package's.

## Photos

6 featured pictures from Wikimedia Commons, 4,003 to 6,329 pixels on the long
side:

| name | photo | by | sha256 of the original file |
|---|---|---|---|
| casting | [Bronze casting at Kunstgießerei München](https://commons.wikimedia.org/wiki/File:Bronze_casting_at_Kunstgie%C3%9Ferei_M%C3%BCnchen_01_-_cropped.jpg) | Kritzolina, CC BY-SA 4.0 | `e4e8572f175566db1cbf36b8597243076b9317c268d4f1bcf2589a42ccd3963d` |
| dawn | [Amanhecer no Hercules](https://commons.wikimedia.org/wiki/File:Amanhecer_no_Hercules_--.jpg) | Carlos Perez Couto, CC BY-SA 3.0 | `ca5f82c44193e99e59ad9f236a23843785e5a9f8cfd61098aa5571d6072d2c92` |
| forest | [Cloud Forest, Gardens by the Bay](https://commons.wikimedia.org/wiki/File:Cloud_Forest,_Gardens_by_the_Bay,_Singapore.jpg) | Basile Morin, CC BY-SA 4.0 | `0dbe6130c88af567377b92f926eb570f0eb035529fea02056eaeb98dd80a7333` |
| night | [Busskur i Riisskov](https://commons.wikimedia.org/wiki/File:Busskur_i_Riisskov_natfoto_version_2.jpg) | Villy Fink Isaksen, CC BY-SA 4.0 | `cb7e0696c4be59e7583c5eb7aa3e74fcfe5cd13a8dcf1ac7300ea203e0f05f62` |
| portrait | [Rainer Pariasek](https://commons.wikimedia.org/wiki/File:20180602_FIFA_Friendly_Match_Austria_vs._Germany_Rainer_Pariasek_850_0540.jpg) | Granada, CC BY-SA 4.0 | `31eaa88a0ddd3cac56442dea696a87003cdbf3eadbfadebb0a0f151e19d048a4` |
| rainbow | [Rainbow over Washfold](https://commons.wikimedia.org/wiki/File:2013_Rainbow_over_Washfold.jpg) | Kreuzschnabel, CC BY-SA 3.0 | `f45bf625f671881b286a47e686855d7bf588b375b552f6d333eaa9ae7488d174` |

None of the photos is included in this repository.

Each was fitted inside 32 and 100 pixels, cropped to 1000x750 and 1920x1080, and
saved by `image/jpeg`. A result is the mean time of 1 encode across all 6.

## Machines and versions

Every implementation ran alone in its own cloud sandbox, once per architecture:

- arm64: Arm Neoverse V2, 1 vCPU, 2 GB RAM
- amd64: 1 vCPU, 2 GB RAM

| implementation | version | how it ran |
|---|---|---|
| this package | this repository | Go 1.27.2 |
| [bbrks/go-blurhash](https://github.com/bbrks/go-blurhash) | v1.2.0 | Go 1.27.2, in the same test binary |
| [buckket/go-blurhash](https://github.com/buckket/go-blurhash) | v1.1.0 | Go 1.27.2, in the same test binary |
| [woltapp/blurhash](https://github.com/woltapp/blurhash), C | commit `dd1986b5` | `C/encode.c`, gcc 14.4.0, `-std=c99 -O2 -ffp-contract=off` |
| [blurhash-python](https://pypi.org/project/blurhash-python/) | 1.2.2 | installed from PyPI, Python 3.12 |
| [woltapp/blurhash](https://www.npmjs.com/package/blurhash), TypeScript | 2.0.5 | installed from npm, Node 22.23.3 |
| [halcy/blurhash-python](https://pypi.org/project/blurhash/) | 1.1.5 | installed from PyPI, Python 3.13.16 |

On the build these tables used, the same code built twice timed up to 0.71%
apart on arm64 and 0.87% on amd64. Every ratio below is 1.45x or more.

## Encoding

From RGBA. This package's time, then every other implementation as a multiple
of it.

**arm64**

| size | components | this package | bbrks | buckket | C | blurhash-python | TypeScript | halcy |
|---|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 4.10 µs | 4.1x | 10.0x | 99.5x | 139.4x | 496.2x | 1,124.5x |
| 32 px | 4x3 | 3.97 µs | 3.2x | 8.5x | 76.9x | 115.7x | 382.6x | 893.8x |
| 100 px | 4x3 | 33.47 µs | 3.1x | 9.1x | 89.9x | 131.4x | 432.7x | 1,026.8x |
| 100 px | 9x9 | 54.03 µs | 12.7x | 26.7x | 385.7x | 447.9x | 1,891.2x | 3,913.6x |
| 1000x750 | 4x3 | 3.52 ms | 2.9x | 8.7x | 88.5x | 127.9x | 428.9x | 1,071.3x |
| 1000x750 | 9x9 | 5.23 ms | 13.1x | 27.8x | 407.2x | 471.0x | 2,023.0x | 4,285.9x |
| 1920x1080 | 4x3 | 9.96 ms | 2.8x | 8.5x | 85.8x | 125.8x | - | - |
| 1920x1080 | 9x9 | 14.45 ms | 13.0x | 27.9x | 405.7x | 470.1x | - | - |

**amd64**

| size | components | this package | bbrks | buckket | C | blurhash-python | TypeScript | halcy |
|---|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 5.62 µs | 3.8x | 8.8x | 60.6x | 101.9x | 430.9x | 1,059.6x |
| 32 px | 4x3 | 5.30 µs | 3.1x | 7.7x | 48.4x | 87.3x | 338.3x | 867.8x |
| 100 px | 4x3 | 37.28 µs | 3.7x | 10.0x | 67.8x | 117.9x | 468.5x | 1,219.4x |
| 100 px | 9x9 | 62.55 µs | 14.3x | 26.2x | 281.0x | 372.9x | 1,932.2x | 4,382.5x |
| 1000x750 | 4x3 | 3.54 ms | 3.6x | 10.3x | 73.9x | 127.6x | 500.2x | 1,372.8x |
| 1000x750 | 9x9 | 5.20 ms | 16.6x | 33.1x | 345.3x | 457.0x | 2,348.6x | 5,524.2x |
| 1920x1080 | 4x3 | 9.73 ms | 3.6x | 10.4x | 73.7x | 128.9x | - | - |
| 1920x1080 | 9x9 | 14.29 ms | 16.7x | 34.4x | 344.9x | 459.8x | - | - |

## Encoding YCbCr

From `*image.YCbCr` at 4:2:0, as `image/jpeg` returns it.

| size | components | arm64: this package | bbrks | buckket | amd64: this package | bbrks | buckket |
|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 5.91 µs | 4.2x | 7.6x | 7.54 µs | 4.1x | 7.0x |
| 32 px | 4x3 | 5.78 µs | 3.6x | 6.5x | 7.23 µs | 3.6x | 6.1x |
| 100 px | 4x3 | 50.75 µs | 3.6x | 6.8x | 55.00 µs | 4.2x | 7.3x |
| 100 px | 9x9 | 71.55 µs | 10.7x | 20.4x | 80.17 µs | 12.3x | 20.9x |
| 1000x750 | 4x3 | 5.35 ms | 3.4x | 6.5x | 5.49 ms | 4.0x | 7.3x |
| 1000x750 | 9x9 | 7.01 ms | 10.9x | 21.4x | 7.13 ms | 13.5x | 25.5x |
| 1920x1080 | 4x3 | 15.44 ms | 3.2x | 6.2x | 15.02 ms | 4.1x | 7.3x |
| 1920x1080 | 9x9 | 19.82 ms | 10.6x | 20.6x | 19.63 ms | 13.6x | 25.4x |

## Decoding

The original C code's hashes of the photos, as a multiple of this package's
time.

| size | components | arm64: bbrks | buckket | amd64: bbrks | buckket |
|---|---|---|---|---|---|
| 32 px | 4x4 | 2.09x | 28.7x | 2.59x | 37.6x |
| 32 px | 4x3 | 1.72x | 24.8x | 2.11x | 31.6x |
| 100 px | 4x3 | 1.52x | 24.3x | 1.91x | 31.0x |
| 100 px | 9x9 | 6.13x | 84.4x | 7.95x | 108.6x |
| 1000x750 | 4x3 | 1.45x | 23.7x | 1.89x | 32.2x |
| 1000x750 | 9x9 | 6.11x | 76.3x | 7.80x | 106.8x |

## Agreement with the original C code

On 9,739 hashes of this package's test images, from the same RGB pixels as the
timings, how many each implementation gives exactly as the original C code
does:

| implementation | arm64 | amd64 |
|---|---|---|
| woltapp/blurhash, C | 9,739 | 9,739 |
| blurhash-python | 9,736 | 9,739 |
| halcy/blurhash-python | 9,547 | 9,547 |
| buckket/go-blurhash | 9,513 | 9,517 |
| this package | 9,508 | 9,508 |
| woltapp/blurhash, TypeScript | 9,418 | 9,418 |
| bbrks/go-blurhash | 9,053 | 9,057 |

This counts matches with the original, not correctness. The original matches
itself by definition, and blurhash-python wraps a copy of the same C code, built
differently on arm64. This package's 231 differences are explained in the README
and listed in [testdata/differences.txt](testdata/differences.txt). A different
count on arm64 and amd64 means that implementation gives some images a different
hash on each.

## Reference hashes

`internal/testimage/testdata/comparison.txt` holds the original C code's hash of
every test image. It was made with the same C build as the timings above, which
gives the same hashes on arm64 and amd64.
