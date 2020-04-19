# 🥦 Broccoli
> `go get -u aletheia.icu/broccoli`

[![GoDoc](https://godoc.org/aletheia.icu/broccoli/fs?status.svg)](https://godoc.org/aletheia.icu/broccoli/fs)
[![Travis](https://travis-ci.org/aletheia-icu/broccoli.svg)](https://travis-ci.org/aletheia-icu/broccoli)
[![Go Report Card](https://goreportcard.com/badge/aletheia.icu/broccoli/fs)](https://goreportcard.com/report/aletheia.icu/broccoli/fs)
[![codecov.io](https://codecov.io/gh/aletheia-icu/broccoli/coverage.svg)](https://codecov.io/gh/aletheia-icu/broccoli)

Broccoli uses [brotli](https://github.com/google/brotli) compression to embed a
virtual file system of static files inside Go executables.

A few reasons to pick broccoli over the alternatives:

- ⚡️ The average is 13-25% smaller binary size due to use of superior
compression algorithm, [brotli](https://github.com/google/brotli).
- 💾 Broccoli supports bundling of multiple source directories, only relies on
`go generate` command-line interface and doesn't require configuration files.
- 🔑 Optional decompression is something you may want; when it's enabled, files
are decompressed only when they are read the first time.
- 🚙 You might want to target `wasm/js` architecture.
- 📰 There is `-gitignore` option to ignore files, already ignored by your
existing .gitignore files.

### Performance
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

#### How does it compare to others?
![](https://i.imgur.com/vB9Miae.png)

Broccoli seems to outperform the existing solutions.

We did [benchmarks](https://vcs.aletheia.icu/lads/broccoli-bench), please feel
free to review them and correct us whenever our methodology could be flawed.

### Usage
```
$ broccoli
Usage: broccoli [options]

Broccoli uses brotli compression to embed a virtual file system in Go executables.

Options:
	-src folder[,file,file2]
		The input files and directories, "public" by default.
	-o
		Name of the generated file, follows input by default.
	-var=br
		Name of the exposed variable, "br" by default.
	-include *.html,*.css
		Wildcard for the files to include, no default.
	-exclude *.wasm
		Wildcard for the files to exclude, no default.
	-opt
		Optional decompression: if enabled, files will only be decompressed
		on the first time they are read.
	-gitignore
		Enables .gitignore rules parsing in each directory, disabled by default.
	-quality [level]
		Brotli compression level (1-11), the highest by default.

Generate a broccoli.gen.go file with the variable broccoli:
	//go:generate broccoli -src assets -o broccoli -var broccoli

Generate a regular public.gen.go file, but include all *.wasm files:
	//go:generate broccoli -src public -include="*.wasm"
```

How broccoli is used in the user code:
```go
//go:generate broccoli -src=testdata,others -o assets
func init() {
    br.Walk("testdata", func(path string, info os.FileInfo, err error) error {
        // walk...
        return nil
    })
}

func main() {
    server := http.FileServer(br)
    http.ListenAndServe(":8080", server)
}
```

### Credits
License: [MIT](https://vcs.aletheia.icu/lads/broccoli/src/branch/master/LICENSE)

We would like to thank brotli development team from Google and Andy Balholm, for
his c2go pure-Go port of the library. Broccoli itself is an effort of a mentoring
experiment, lead by [@tucnak](https://github.com/tucnak) on the foundation of
[Aletheia](https://aletheia.icu).
