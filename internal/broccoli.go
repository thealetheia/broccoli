// Package internal implements compression and virtual file system access.
package internal

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/andybalholm/brotli"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	compressQuality = 11
)

var SkipDir = errors.New("skip this directory")

type Broccoli struct {
	files     map[string]*File
	fileNames []string
}

type WalkFunc func(path string, f *File) error

func (br *Broccoli) Walk(root string, walkFn WalkFunc) error {
	pos := sort.SearchStrings(br.fileNames, root)
	for pos < len(br.fileNames) && strings.HasPrefix(br.fileNames[pos], root) {
		if br.files[br.fileNames[pos]].IsDir() == false {
			err := walkFn(br.fileNames[pos], br.files[br.fileNames[pos]])
			if err != nil {
				log.Fatal(err)
				return err
			}
		}
		pos++
	}

	return nil
}

func (br *Broccoli) Stat(filepath string) (os.FileInfo, error) {
	fileInfo, err := os.Lstat(filepath)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func (br *Broccoli) Open(filepath string) (*File, error) {
	butchered := strings.Split(filepath, "/")
	name := butchered[len(butchered)-1]
	return NewFile(filepath, name)
}

func (br *Broccoli) uncompress(data []byte) ([]File, error) {
	var filesCount uint
	decodeBuffer := bytes.NewBuffer(data)
	brotliReader := brotli.NewReader(decodeBuffer)
	n, err := brotliReader.Read(data)

	if err != nil {
		log.Fatal(err)
		return []File{}, err
	}

	buffer := bytes.NewBuffer(decodeBuffer.Bytes())
	buffer.Grow(n)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&filesCount)
	if err != nil {
		log.Fatal(err)
		return []File{}, err
	}
	files := make([]File, filesCount)
	for i := uint(0); i < filesCount; i++ {
		err := decoder.Decode(&files[i])
		if err != nil {
			log.Fatal(err)
			return files, err
		}
	}

	return files, nil

}

func Pack(files []*File) ([]byte, error) {
	var bundle []byte
	for _, file := range files {
		compressedFile, err := file.compress()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		bundle = append(bundle, compressedFile...)
	}
	return bundle, nil
}

func New(bundle []byte) *Broccoli {
	return nil
}

type File struct {
	size         int64
	timeModified int64
	name         string
	filepath     string
	data         []byte
	compressed   bool
	fd           int64
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Size() int64 {
	return f.size
}

func (f *File) Mode() os.FileMode {
	if f.IsDir() == false {
		return 0444
	} else {
		return os.ModeDir
	}
}

func (f *File) ModTime() time.Time {
	return time.Unix(f.timeModified, 0)
}

func (f *File) IsDir() bool {
	return f.data == nil
}

func (f *File) Sys() interface{} {
	return nil
}
func (f *File) Read(b []byte) (int, error) {
	r, err := syscall.Open(f.Name(), syscall.O_CLOEXEC, uint32(os.FileMode(0444)))
	f.fd = int64(r)
	if err != nil {
		return 0, err
	}
	return syscall.Read(r, b)
}
func (f *File) Close() error {
	if f.fd < 0 {
		return errors.New("[ERROR] Already closed")
	}
	e := syscall.Close(int(f.fd))
	if e != nil {
		return e
	}
	return nil

}

func NewFile(filepath, name string) (*File, error) {
	fileInfo, err := os.Stat(filepath)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	fileBytes, err := readBytes(filepath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &File{fileInfo.Size(), fileInfo.ModTime().Unix(), name,
		filepath, fileBytes, false, -1}, nil
}

func (f *File) compress() ([]byte, error) {
	var gobEncoding bytes.Buffer
	enc := gob.NewEncoder(&gobEncoding)
	err := enc.Encode(f)

	if err != nil {
		log.Fatal(err)
		return []byte{}, err
	}

	var compressedBuffer bytes.Buffer
	compressedBuffer.Grow(len(gobEncoding.Bytes()))
	compressedWriter := brotli.NewWriterLevel(&compressedBuffer, compressQuality)
	defer compressedWriter.Close()
	_, err = compressedWriter.Write(gobEncoding.Bytes())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return compressedBuffer.Bytes(), nil
}

func readBytes(filepath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return file, nil
}
