package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	br "aletheia.icu/broccoli/broccoli"
)

func TestBroccoli(t *testing.T) {
	var (
		realPaths    []string
		virtualPaths []string

		files []*br.File
	)

	filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		f, err := br.NewFile(path)
		if err != nil {
			t.Fatal(err)
		}

		files = append(files, f)
		realPaths = append(realPaths, f.Fpath)
		return nil
	})

	bytes, err := br.Pack(files, 11)
	if err != nil {
		t.Fatal(err)
	}

	br, err := br.New(bytes)
	if err != nil {
		t.Fatal(err)
	}

	br.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	if !assert.Equal(t, realPaths, virtualPaths, "paths asymmetric") {
		t.Fatal()
	}
}
