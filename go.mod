module broccoli

go 1.13

require (
	github.com/andybalholm/brotli v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1

	aletheia.icu/broccoli/broccoli v0.0.0
)

replace aletheia.icu/broccoli/broccoli => ./broccoli
