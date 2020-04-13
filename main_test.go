package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"aletheia.icu/broccoli/fs"
)

var (
	bundle, _ = defaultGenerator().generate()
	br        = fs.New(false, bundle)
)

func defaultGenerator() *Generator {
	return &Generator{
		inputFiles: []string{"testdata"},
		quality:    11,
	}
}

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

	br := fs.New(false, bundle)
	br.Walk("./testdata", func(path string, _ os.FileInfo, _ error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	assert.Equal(t, realPaths, virtualPaths, "paths asymmetric")
	fmt.Printf("testdata: compression factor %.2fx\n", totalSize/float64(len(bundle)))
}

func TestGenerate(t *testing.T) {
	walk := func(g *Generator, walkFn filepath.WalkFunc) {
		bundle, err := g.generate()
		if err != nil {
			t.Fatal(err)
		}

		br := fs.New(false, bundle)
		br.Walk("testdata", walkFn)
	}

	var (
		realPaths    []string
		virtualPaths []string
	)

	filepath.Walk("testdata", func(path string, _ os.FileInfo, _ error) error {
		path = strings.ReplaceAll(path, `\`, "/")
		realPaths = append(realPaths, path)
		return nil
	})

	g := defaultGenerator()
	walk(g, func(path string, _ os.FileInfo, _ error) error {
		virtualPaths = append(virtualPaths, path)
		return nil
	})

	// to be sure that generator without side-effects gives exactly the same file structure
	assert.Equal(t, realPaths, virtualPaths, "paths asymmetric")

	g = defaultGenerator()
	g.includeGlob = "*.html"
	walk(g, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && filepath.Ext(path) != ".html" {
			t.Fatalf("generated bundle should not include excluded files")
		}
		return nil
	})

	g = defaultGenerator()
	g.excludeGlob = "*.html"
	walk(g, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			t.Fatalf("generated bundle should not include excluded files")
		}
		return nil
	})

	g = defaultGenerator()
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

func TestFile(t *testing.T) {
	_, err := fs.NewFile("bad")
	assert.Error(t, err)

	f, err := br.Open("testdata/index.html")
	assert.NoError(t, err)
	_, ok := f.(*fs.File)
	assert.True(t, ok)

	info, err := os.Stat("testdata/index.html")
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0444), f.(*fs.File).Mode()) // const for files
	assert.Equal(t, info.ModTime().Truncate(time.Second), f.(*fs.File).ModTime())

	stat, err := f.Stat()
	assert.NoError(t, err)
	assert.NotNil(t, stat)

	assert.NoError(t, f.Close())
	_, err = f.Read(nil)
	assert.Equal(t, os.ErrClosed, err)
	_, err = f.Readdir(0)
	assert.Error(t, os.ErrInvalid, err)
	assert.Equal(t, os.ErrClosed, f.Close())

	assert.Equal(t, "index.html", f.(*fs.File).Name())
	assert.Equal(t, info.Size(), f.(*fs.File).Size())
	assert.False(t, f.(*fs.File).IsDir())
	assert.Nil(t, f.(*fs.File).Sys())

	assert.NoError(t, f.(*fs.File).Open())
	_, err = f.Read(make([]byte, 0, 32))
	assert.NoError(t, err)

	dir, err := br.Open("testdata/html")
	assert.NoError(t, err)

	info, err = os.Stat("testdata/html")
	assert.NoError(t, err)
	assert.Equal(t, os.ModeDir, dir.(*fs.File).Mode())
	assert.Equal(t, info.ModTime().Truncate(time.Second), dir.(*fs.File).ModTime())
}

func TestFileReaddir(t *testing.T) {
	g := Generator{
		inputFiles: []string{"testdata/readdir"},
		quality:    11,
	}

	bundle, err := g.generate()
	if err != nil {
		t.Fatal(err)
	}
	br := fs.New(false, bundle)

	dir, err := br.Open("testdata/readdir")
	if err != nil {
		t.Fatal(err)
	}

	infos, err := dir.Readdir(-1)
	assert.NoError(t, err)
	assert.Len(t, infos, 3)

	infos, err = dir.Readdir(0)
	assert.NoError(t, err)
	assert.Len(t, infos, 3)

	infos, err = dir.Readdir(1)
	assert.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, "1.txt", infos[0].Name())

	infos, err = dir.Readdir(1)
	assert.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, "2.txt", infos[0].Name())

	infos, err = dir.Readdir(1)
	assert.NoError(t, err)
	assert.Len(t, infos, 1)
	assert.Equal(t, "3.txt", infos[0].Name())

	_, err = dir.Readdir(1)
	assert.Error(t, err)
}

func TestFileSeek(t *testing.T) {
	// TODO
}

func TestHttpFileServer(t *testing.T) {
	srv := httptest.NewServer(http.FileServer(br))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/testdata/index.html")
	assert.NoError(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	orig, err := ioutil.ReadFile("testdata/index.html")
	assert.NoError(t, err)

	assert.Equal(t, data, orig)
	t.Log(string(data))
}
