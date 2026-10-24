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

- arm64: Arm Neoverse V2, 4 vCPU, 8 GiB RAM
- amd64: 4 vCPU, 8 GiB RAM

Each run was limited to 2 of the 4 CPUs.

| implementation | version | how it ran |
|---|---|---|
| this package | this repository | Go 1.27.2 |
| [bbrks/go-blurhash](https://github.com/bbrks/go-blurhash) | v1.2.0 | Go 1.27.2, in the same test binary |
| [buckket/go-blurhash](https://github.com/buckket/go-blurhash) | v1.1.0 | Go 1.27.2, in the same test binary |
| [woltapp/blurhash](https://github.com/woltapp/blurhash), C | commit `dd1986b5` | `C/encode.c`, gcc 14.4.0, `-std=c99 -O2 -ffp-contract=off` |
| [blurhash-python](https://pypi.org/project/blurhash-python/) | 1.2.2 | installed from PyPI, Python 3.12 |
| [woltapp/blurhash](https://www.npmjs.com/package/blurhash), TypeScript | 2.0.5 | installed from npm, Node 22.23.3 |
| [halcy/blurhash-python](https://pypi.org/project/blurhash/) | 1.1.5 | installed from PyPI, Python 3.13.16 |

## Encoding

From RGBA. This package's time, then every other implementation as a multiple
of it.

**arm64**

| size | components | this package | bbrks | buckket | C | blurhash-python | TypeScript | halcy |
|---|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 4.09 µs | 4.0x | 10.0x | 99.4x | 139.9x | 492.1x | 1,131.2x |
| 32 px | 4x3 | 3.96 µs | 3.2x | 8.5x | 76.9x | 116.1x | 376.3x | 896.0x |
| 100 px | 4x3 | 33.44 µs | 3.1x | 9.1x | 90.0x | 131.7x | 431.0x | 1,026.6x |
| 100 px | 9x9 | 53.68 µs | 12.7x | 26.9x | 387.5x | 451.4x | 1,902.3x | 3,935.4x |
| 1000x750 | 4x3 | 3.53 ms | 2.9x | 8.7x | 87.9x | 127.5x | 427.7x | 1,069.0x |
| 1000x750 | 9x9 | 5.22 ms | 13.1x | 28.0x | 406.8x | 472.4x | 2,027.5x | 4,312.7x |
| 1920x1080 | 4x3 | 10.04 ms | 2.8x | 8.5x | 84.9x | 124.9x | - | - |
| 1920x1080 | 9x9 | 14.46 ms | 13.0x | 27.9x | 403.9x | 469.8x | - | - |

**amd64**

| size | components | this package | bbrks | buckket | C | blurhash-python | TypeScript | halcy |
|---|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 5.56 µs | 3.9x | 8.8x | 61.3x | 102.6x | 429.5x | 1,073.5x |
| 32 px | 4x3 | 5.27 µs | 3.1x | 7.8x | 48.5x | 87.6x | 336.2x | 878.1x |
| 100 px | 4x3 | 36.84 µs | 3.8x | 10.1x | 68.7x | 119.8x | 468.2x | 1,212.3x |
| 100 px | 9x9 | 62.14 µs | 14.5x | 26.4x | 282.9x | 377.4x | 1,945.2x | 4,407.6x |
| 1000x750 | 4x3 | 3.49 ms | 3.7x | 10.6x | 74.7x | 129.3x | 505.5x | 1,400.4x |
| 1000x750 | 9x9 | 5.17 ms | 16.7x | 33.5x | 345.3x | 464.1x | 2,358.6x | 5,543.0x |
| 1920x1080 | 4x3 | 9.68 ms | 3.6x | 10.7x | 74.1x | 129.8x | - | - |
| 1920x1080 | 9x9 | 14.20 ms | 16.9x | 35.1x | 345.9x | 462.5x | - | - |

## Encoding YCbCr

From `*image.YCbCr` at 4:2:0, as `image/jpeg` returns it.

| size | components | arm64: this package | bbrks | buckket | amd64: this package | bbrks | buckket |
|---|---|---|---|---|---|---|---|
| 32 px | 4x4 | 5.90 µs | 4.1x | 7.6x | 7.63 µs | 4.1x | 6.9x |
| 32 px | 4x3 | 5.77 µs | 3.6x | 6.5x | 7.34 µs | 3.5x | 6.1x |
| 100 px | 4x3 | 51.08 µs | 3.5x | 6.7x | 55.63 µs | 4.2x | 7.2x |
| 100 px | 9x9 | 71.20 µs | 10.7x | 20.6x | 80.88 µs | 12.4x | 20.7x |
| 1000x750 | 4x3 | 5.34 ms | 3.4x | 6.5x | 5.52 ms | 4.1x | 7.3x |
| 1000x750 | 9x9 | 7.01 ms | 10.9x | 21.2x | 7.22 ms | 13.3x | 24.8x |
| 1920x1080 | 4x3 | 15.45 ms | 3.2x | 6.2x | 15.17 ms | 4.1x | 7.3x |
| 1920x1080 | 9x9 | 19.82 ms | 10.6x | 20.7x | 19.94 ms | 13.3x | 25.8x |

## Decoding

The original C code's hashes of the photos. This package's time, then every
other implementation as a multiple of it.

**arm64**

| size | components | this package | bbrks | buckket |
|---|---|---|---|---|
| 32 px | 4x4 | 9.78 µs | 2.07x | 28.7x |
| 32 px | 4x3 | 9.56 µs | 1.71x | 24.8x |
| 100 px | 4x3 | 97.01 µs | 1.51x | 24.1x |
| 100 px | 9x9 | 119.87 µs | 6.07x | 83.5x |
| 1000x750 | 4x3 | 10.06 ms | 1.43x | 23.7x |
| 1000x750 | 9x9 | 11.96 ms | 6.11x | 76.0x |

**amd64**

| size | components | this package | bbrks | buckket |
|---|---|---|---|---|
| 32 px | 4x4 | 11.37 µs | 2.73x | 37.3x |
| 32 px | 4x3 | 11.19 µs | 2.19x | 31.8x |
| 100 px | 4x3 | 112.57 µs | 1.97x | 30.4x |
| 100 px | 9x9 | 137.02 µs | 8.48x | 107.0x |
| 1000x750 | 4x3 | 10.93 ms | 1.97x | 32.2x |
| 1000x750 | 9x9 | 13.85 ms | 8.46x | 105.3x |

## Agreement with the original C code

On 9,719 hashes of this package's test images, from the same RGB pixels as the
timings, how many each implementation gives exactly as the original C code
does:

| implementation | arm64 | amd64 |
|---|---|---|
| woltapp/blurhash, C | 9,719 | 9,719 |
| blurhash-python | 9,716 | 9,719 |
| halcy/blurhash-python | 9,527 | 9,527 |
| buckket/go-blurhash | 9,493 | 9,497 |
| this package | 9,488 | 9,488 |
| woltapp/blurhash, TypeScript | 9,403 | 9,403 |
| bbrks/go-blurhash | 9,033 | 9,037 |

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
