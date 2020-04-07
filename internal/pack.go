package broccoli

import (
	"bytes"
	"encoding/binary"
	"sort"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

const packingQuality = 6

func Pack(files []*File, quality int) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Fpath < files[j].Fpath
	})

	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, packingQuality)

	err := binary.Write(w, binary.LittleEndian, uint32(len(files)))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		compressed, err := file.compress(quality)
		if err != nil {
			return nil, errors.Wrap(err, "could not compress "+file.Fpath)
		}

		err = binary.Write(w, binary.LittleEndian, uint64(len(compressed)))
		if err != nil {
			return nil, err
		}

		_, err = w.Write(compressed)
		if err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func New(bundle []byte) (*Broccoli, error) {
	r := brotli.NewReader(bytes.NewReader(bundle))

	var n uint32
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		return nil, err
	}

	br := &Broccoli{
		files:     map[string]*File{},
		filePaths: make([]string, 0, n),
	}

	for i := 0; i < int(n); i++ {
		var m uint64
		err = binary.Read(r, binary.LittleEndian, &m)

		b := make([]byte, m)
		if _, err = r.Read(b); err != nil {
			return nil, err
		}

		f := &File{compressed: true}
		if err := f.decompress(b); err != nil {
			return nil, err
		}

		br.files[f.Fpath] = f
		br.filePaths = append(br.filePaths, f.Fpath)
	}

	return br, nil
}
