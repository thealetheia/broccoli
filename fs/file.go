package fs

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
)

var root string

func init() {
	root, _ = os.Getwd()
}

// File represents a bundled asset.
//
// It should never be created explicitly, but rather accessed
// via Open(), as it only makes sense to create it in the
// context of the broccoli tool itself.
type File struct {
	compressed bool

	Data  []byte
	Fpath string
	Fname string
	Fsize int64
	Ftime int64

	buffer *bytes.Buffer
	br     *Broccoli
}

// Stat returns a FileInfo describing this file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.br.Stat(f.Fpath)
}

// NewFile constructs a new bundled file from the disk.
//
// It is only supposed to be called from the broccoli tool.
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

// Open opens the file for reading. If successful, methods on
// the returned file can be used for reading.
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

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (f *File) Read(b []byte) (int, error) {
	if f.buffer == nil {
		return 0, os.ErrClosed
	}

	return f.buffer.Read(b)
}

var (
	errBadOffset = errors.New("Seek: bad offset")
	errBadWhence = errors.New("Seek: bad whence")
)

// Seek sets the offset for the next Read or Write on file to offset,
// interpreted according to whence: 0 means relative to the origin of the file,
// 1 means relative to the current offset, and 2 means relative to the end.
//
// It returns the new offset and and error, if any.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.buffer == nil {
		return 0, os.ErrClosed
	}

	n := int64(len(f.Data))

	switch whence {
	// io.SeekStart
	// seek relative to the origin of the file
	case 0:
		if offset >= n {
			return 0, errBadOffset
		}
		f.buffer = bytes.NewBuffer(f.Data[offset:])
		return offset, nil
	// io.SeekCurrent
	// seek relative to the current offset
	case 1:
		if offset >= int64(f.buffer.Len()) {
			return 0, errBadOffset
		}
		i := n - int64(f.buffer.Len()) + offset
		f.buffer = bytes.NewBuffer(f.Data[i:])
		return i, nil
	// io.SeekEnd
	// seek relative to the end
	case 2:
		if offset >= n {
			return 0, errBadOffset
		}

		i := n - offset
		f.buffer = bytes.NewBuffer(f.Data[i:])
		return i, nil
	default:
		return 0, errBadWhence
	}
}

// Close clears the dedicated file buffer.
func (f *File) Close() error {
	if f.buffer == nil {
		return os.ErrClosed
	}

	f.buffer = nil
	return nil
}

// Name returns the basename of the file.
func (f *File) Name() string {
	return f.Fname
}

// Size returns the size of the file in bytes.
func (f *File) Size() int64 {
	return f.Fsize
}

// Mode returns the file mode of the file.
//
// It's os.ModeDir for directories, 0444 otherwise.
func (f *File) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}

	return 0444
}

// ModTime returns the time file was last modified.
func (f *File) ModTime() time.Time {
	t := f.Ftime
	if t < 0 {
		t = -t
	}
	return time.Unix(t, 0)
}

// IsDir tells whether if the file is a directory.
func (f *File) IsDir() bool {
	return f.Ftime < 0
}

// Readdir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values, as would be returned
// by Lstat, in directory order. Subsequent calls on the same file will yield
// further FileInfos.
//
// If n > 0, Readdir returns at most n FileInfo structures. In this case, if
// Readdir returns an empty slice, it will return a non-nil error
// explaining why. At the end of a directory, the error is io.EOF.
//
// If n <= 0, Readdir returns all the FileInfo from the directory in
// a single slice. In this case, if Readdir succeeds (reads all
// the way to the end of the directory), it returns the slice and a
// nil error. If it encounters an error before the end of the
// directory, Readdir returns the FileInfo read until that point
// and a non-nil error.
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	if !f.IsDir() {
		return nil, os.ErrInvalid
	}

	if count < 0 {
		count = 0
	}

	files := make([]os.FileInfo, 0, count)
	for i := sort.SearchStrings(f.br.filePaths, f.Fpath) + 1; ; i++ {
		g := f.br.files[f.br.filePaths[i]]
		if !strings.HasPrefix(g.Fpath, f.Fpath) {
			break
		}

		files = append(files, g)
		if count != 0 && len(files) == count {
			break
		}
	}

	return files, nil
}

// Sys is a mystery and always returns nil.
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
	f.compressed = false
	return nil
}
