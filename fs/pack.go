package fs

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"sort"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

const packingQuality = 6

func Pack(files []*File, quality int) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Fpath < files[j].Fpath
	})

	for _, file := range files {
		data, err := file.compress(quality)
		if err != nil {
			return nil, err
		}

		file.Data = data
	}

	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, packingQuality)

	if err := gob.NewEncoder(w).Encode(files); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return encode(b.Bytes()), nil
}

func New(payload []byte) *Broccoli {
	bundle := decode(payload)

	var files []*File
	r := brotli.NewReader(bytes.NewBuffer(bundle))
	if err := gob.NewDecoder(r).Decode(&files); err != nil {
		panic(err)
	}

	br := &Broccoli{
		filePaths: make([]string, 0, len(files)),
		files:     map[string]*File{},
	}

	for _, file := range files {
		if err := file.decompress(file.Data); err != nil {
			panic(errors.Wrap(err, "could not decompress file"))
		}

		br.files[file.Fpath] = file
		br.filePaths = append(br.filePaths, file.Fpath)
	}

	return br
}

func encode(b []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(b))
}

func decode(b []byte) []byte {
	d, _ := base64.StdEncoding.DecodeString(string(b))
	return d
}
