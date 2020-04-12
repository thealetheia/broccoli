package fs

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

var root string

func init() {
	root, _ = os.Getwd()
}

type File struct {
	compressed bool

	Data  []byte
	Fpath string
	Fname string
	Fsize int64
	Ftime int64

	buffer *bytes.Buffer
}

// NewFile is only supposed to be called from package main.
func NewFile(path string) (*File, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil && !fileInfo.IsDir() {
		return nil, err
	}

	path, _ = filepath.Abs(path)
	path, _ = filepath.Rel(root, path)

	// NOTE: On Windows, it evidently does happen.
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, `\`, "/")
	}

	time := fileInfo.ModTime().Unix()
	if fileInfo.IsDir() {
		time = -time
	}

	return &File{
		Data:  data,
		Fpath: path,
		Fname: fileInfo.Name(),
		Fsize: fileInfo.Size(),
		Ftime: time,
	}, nil
}

func (f *File) Open() error {
	if f.IsDir() {
		return os.ErrPermission
	}

	if f.compressed {
		if err := f.decompress(f.Data); err != nil {
			return errors.Wrap(err, "could not decompress")
		}
	}

	f.buffer = bytes.NewBuffer(f.Data)
	return nil
}

func (f *File) Read(b []byte) (int, error) {
	if f.buffer == nil {
		return 0, os.ErrClosed
	}

	return f.buffer.Read(b)
}

func (f *File) Close() error {
	if f.buffer == nil {
		return os.ErrClosed
	}

	f.buffer = nil
	return nil
}

func (f *File) Name() string {
	return f.Fname
}

func (f *File) Size() int64 {
	return f.Fsize
}

func (f *File) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	} else {
		return 0444
	}
}

func (f *File) ModTime() time.Time {
	t := f.Ftime
	if t < 0 {
		t = -t
	}
	return time.Unix(t, 0)
}

func (f *File) IsDir() bool {
	return f.Ftime < 0
}

func (f *File) Sys() interface{} {
	return nil
}

func (f *File) compress(quality int) ([]byte, error) {
	var b bytes.Buffer
	w := brotli.NewWriterLevel(&b, quality)
	_, err := w.Write(f.Data)
	if err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (f *File) decompress(data []byte) error {
	r := brotli.NewReader(bytes.NewReader(data))
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f.Data = b
	return nil
}
