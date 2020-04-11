// Package fs implements virtual file system access for compressed files.
package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Broccoli struct {
	files     map[string]*File
	filePaths []string
}

func (br *Broccoli) Open(filepath string) (*File, error) {
	filepath = normalize(filepath)

	if file, ok := br.files[filepath]; ok {
		if err := file.Open(); err != nil {
			return nil, err
		}
		return file, nil
	} else {
		return nil, os.ErrNotExist
	}
}

func (br *Broccoli) Stat(path string) (os.FileInfo, error) {
	path = normalize(path)

	if file, ok := br.files[path]; ok {
		return file, nil
	} else {
		return nil, os.ErrNotExist
	}
}

func (br *Broccoli) Walk(root string, walkFn filepath.WalkFunc) error {
	root = normalize(root)

	pos := sort.SearchStrings(br.filePaths, root)
	for ; pos < len(br.filePaths) && strings.HasPrefix(br.filePaths[pos], root); pos++ {
		file := br.files[br.filePaths[pos]]
		err := walkFn(file.Fpath, file, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func normalize(path string) string {
	if strings.HasPrefix(path, "./") {
		return path[2:]
	}

	return path
}
