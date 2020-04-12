# Broccoli
> `go get -u aletheia.icu/broccoli`

[![GoDoc](https://godoc.org/aletheia.icu/broccoli/fs?status.svg)](https://godoc.org/aletheia.icu/broccoli/fs)
[![Go Report Card](https://goreportcard.com/badge/aletheia.icu/broccoli/fs)](https://goreportcard.com/report/aletheia.icu/broccoli/fs)

Broccoli uses [brotli](https://github.com/google/brotli) compression to embed a 
virtual file system of static files inside Go executables.

A few reasons to pick broccoli over the alternatives:

- ⚡️ The average is 13-25% smaller binary size due to use of superior
compression algorithm, [brotli](https://github.com/google/brotli).
- 💾 Broccoli supports bundling of multiple source directories, only relies on
`go generate` command-line interface and doesn't require configuration files.
- 🔑 Optional decompression is something you may want to you; when it's enabled,
files are decompressed only when they are read the first time.
- 🚙 You might want to target `wasm/js` architecture.
- 📰 There is `-gitignore` option to ignore files, already ignored by your
existing .gitignore files.

Admittedly, there are already many packages providing similar functionality out
there in the wild. Tim Shannon did an overall pretty good overview of them in
[Choosing A Library to Embed Static Assets in Go](https://tech.townsourced.com/post/embedding-static-files-in-go/),
but it should be outdated by at least two years, so although we subscribe to the
analysis, we cannot guarantee whether if it's up–to–date. Most if not all of the
packages mentioned in the article, rely on gzip compression and most of them,
unfortunately are not compatible with `wasm/js` architecture, due to some quirk
that has to do with their use of `http` package. This, among other things, was
the driving force behind the creation of broccoli.

The most feature-complete library from the comparison table seems to be
[fileb0x](https://github.com/UnnoTed/fileb0x).

#### How does broccoli compare to flexb0x?
Feature                               | fileb0x             | broccoli
---------------------                 | -----------         | ------------------
compression                           | gzip                | brotli (-20% avg.)
optional decompression                | yes                 | yes
compression levels                    | yes                 | yes (1-11)
different build tags for each file    | yes                 | no
exclude / ignore files                | glob                | glob
unexported vars/funcs                 | optional            | optional
virtual memory file system            | yes                 | yes
http file system                      | yes                 | yes
replace text in files                 | yes                 | no
glob support                          | yes                 | yes
regex support                         | no                  | no
config file                           | yes                 | no
update files remotely                 | yes                 | no
.gitignore support                    | no                  | yes

#### How does it compare to other packages?

Broccoli is supposed to be used with [go generate](https://blog.golang.org/generate).
