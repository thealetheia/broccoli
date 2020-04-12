package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"aletheia.icu/broccoli/fs"
)

func TestBroccoli(t *testing.T) {
	var (
		realPaths    []string
		virtualPaths []string
		totalSize    float64

		files []*fs.File
	)

	filepath.Walk("testdata", func(path string, info os.FileInfo, _ error) error {
		f, err := fs.NewFile(path)
		if err != nil {
			t.Fatal(err)
		}

		files = append(files, f)
		realPaths = append(realPaths, f.Fpath)
		totalSize += float64(info.Size())
		return nil
	})

	bundle, err := fs.Pack(files, 11)
	if err != nil {
		t.Fatal(err)
	}

	br := fs.New(bundle)
	br.Walk("./testdata", func(path string, _ os.FileInfo, _ error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	assert.Equal(t, realPaths, virtualPaths, "paths asymmetric")
	fmt.Printf("testdata: compression factor %.2fx\n", totalSize/float64(len(bundle)))
}

func TestGenerate(t *testing.T) {
	walk := func(g Generator, walkFn filepath.WalkFunc) {
		bundle, err := g.generate()
		if err != nil {
			t.Fatal(err)
		}

		br := fs.New(bundle)
		br.Walk("testdata", walkFn)
	}

	generator := func() Generator {
		return Generator{
			inputFiles:   []string{"testdata"},
			quality:      11,
		}
	}

	var (
		realPaths []string
		virtualPaths []string
	)

	filepath.Walk("testdata", func(path string, _ os.FileInfo, _ error) error {
		path = strings.ReplaceAll(path, `\`, "/")
		realPaths = append(realPaths, path)
		return nil
	})

	g := generator()
	walk(g, func(path string, _ os.FileInfo, _ error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	// to be sure that generator without side-effects gives exactly the same file structure
	assert.Equal(t, realPaths, virtualPaths, "paths asymmetric")

	g = generator()
	g.includeGlob = "*.html"
	walk(g, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && filepath.Ext(path) != ".html" {
			t.Fatalf("generated bundle should not include excluded files")
		}
		return nil
	})

	g = generator()
	g.excludeGlob = "*.html"
	walk(g, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			t.Fatalf("generated bundle should not include excluded files")
		}
		return nil
	})

	g = generator()
	g.useGitignore = true
	walk(g, func(path string, info os.FileInfo, _ error) error {
		// following .gitignore rules
		if info.IsDir() {
			if info.Name() == "contents" {
				t.Fatal("generated bundle contains excluded contents directory")
			}
		} else if info.Name() != ".gitignore" && info.Name() != "googleJS.js" {
			t.Fatal("generated bundle should include only googleJS.js file")
		}
		return nil
	})
}
