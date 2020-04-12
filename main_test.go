package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	fs "aletheia.icu/broccoli/fs"
)

func TestBroccoli(t *testing.T) {
	var (
		realPaths    []string
		virtualPaths []string
		totalSize    float64

		files []*fs.File
	)

	filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		totalSize += float64(info.Size())

		f, err := fs.NewFile(path)
		if err != nil {
			t.Fatal(err)
		}

		files = append(files, f)
		realPaths = append(realPaths, f.Fpath)
		return nil
	})

	bundle, err := fs.Pack(files, 11)
	if err != nil {
		t.Fatal(err)
	}

	br := fs.New(bundle)
	_ = br.Walk("./testdata", func(path string, info os.FileInfo, err error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	if !assert.Equal(t, realPaths, virtualPaths, "paths asymmetric") {
		t.Fatal()
	}

	fmt.Printf("testdata: compression factor %.2fx\n", totalSize/float64(len(bundle)))
}
