package fs

import (
	"bytes"
	"encoding/gob"
	"runtime"
	"sort"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

// Pack compresses a set of files from disk for bundled use in the generated code.
//
// This function is only supposed to be called by broccoli the tool.
func Pack(files []*File, quality int) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Fpath < files[j].Fpath
	})

	n := runtime.NumCPU()
	feed := make(chan *File)
	errs := make(chan error, n)
	defer close(errs)

	for i := 0; i < n; i++ {
		go func() {
			for f := range feed {
				data, err := f.compress(quality)
				if err != nil {
					errs <- err
					return
				}

				f.Data = data
			}

			errs <- nil
		}()
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		feed <- file
	}
	close(feed)

	for err := range errs {
		if err != nil {
			return nil, err
		}

		n--
		if n == 0 {
			break
		}
	}

	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, quality)
	if err := gob.NewEncoder(w).Encode(files); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// New decompresses the bundle byte-slice and creates a virtual file system.
// Depending on whether if optional decompression is enabled, it will or
// will not decompress the files while loading them.
//
// This function is only supposed to be called from the generated code.
func New(opt bool, bundle []byte) *Broccoli {
	var files []*File
	r := brotli.NewReader(bytes.NewBuffer(bundle))
	if err := gob.NewDecoder(r).Decode(&files); err != nil {
		panic(err)
	}

	br := &Broccoli{
		filePaths: make([]string, 0, len(files)),
		files:     map[string]*File{},
	}

	for _, f := range files {
		f.compressed = true
		f.br = br

		br.files[f.Fpath] = f
		br.filePaths = append(br.filePaths, f.Fpath)
	}

	if opt {
		return br
	}

	n := runtime.NumCPU()
	feed := make(chan *File)
	done := make(chan struct{}, n)
	defer close(done)

	for i := 0; i < n; i++ {
		go func() {
			for f := range feed {
				if err := f.decompress(f.Data); err != nil {
					panic(errors.Wrap(err, "could not decompress"))
				}
			}

			done <- struct{}{}
		}()
	}

	for _, file := range files {
		feed <- file
	}
	close(feed)

	for range done {
		n--
		if n == 0 {
			break
		}
	}

	return br
}
