# Broccoli
> go get -u aletheia.icu/broccoli

Broccoli uses [brotli](https://github.com/google/brotli) compression to embed a virtual file system of static files inside Go executables. There are many packages providing similar functionality, but most if not all of them rely on gzip for compression and most of them are not compatible with `wasm/js` architecture, due to their reliance on `http` package.

Broccoli is supposed to be used with [go generate](https://blog.golang.org/generate).

On average, it should perform better than the alternatives.
