# Broccoli
> `go get -u aletheia.icu/broccoli`

[![GoDoc](https://godoc.org/aletheia.icu/broccoli/fs?status.svg)](https://godoc.org/aletheia.icu/broccoli/fs)
[![Go Report Card](https://goreportcard.com/badge/aletheia.icu/broccoli/fs)](https://goreportcard.com/report/aletheia.icu/broccoli/fs)

Broccoli uses [brotli](https://github.com/google/brotli) compression to embed a 
virtual file system of static files inside Go executables.

Admittedly, many packages providing similar functionality already exist. Tim
Shannon did a good overview of them in
[Choosing A Library to Embed Static Assets in Go](https://tech.townsourced.com/post/embedding-static-files-in-go/),
but it should be outdated by at least two years, so although I subscribe to the
analysis, we cannot guarantee it's up–to–date. Most, if not all of the packages
mentioned in the aforementioned piece rely on gzip for compression and most of
them unfortunately are not compatible with `wasm/js` architecture, due to their
implicit reliance on `http` package. This, among other things, was the driving
reason behind the creation of broccoli.

A few reasons to pick broccoli over the alternatives:

- ⚡️ On the average, noticeably smaller binary size due to use of superior
compression algorithm, [brotli](https://github.com/google/brotli).
- 🚙 You target `wasm/js` architecture.
- 🗂 Broccoli supports bundling of multiple source directories, only relies on CLI
interface, accessible via `go generate` and doesn't require configuration files.
- 📰 There is `-gitignore` option to ignore files already ignored by your
existing .gitignore files.
- 🗝 Optional decompression, where files are only decompressed when they are
first read.

Broccoli is supposed to be used with [go generate](https://blog.golang.org/generate).
