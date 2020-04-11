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
	if file, ok := br.files[filepath]; ok {
		if err := file.Open(); err != nil {
			return nil, err
		}
		return file, nil
	} else {
		return nil, os.ErrNotExist
	}
}

func (br *Broccoli) Stat(filepath string) (os.FileInfo, error) {
	if file, ok := br.files[filepath]; ok {
		return file, nil
	} else {
		return nil, os.ErrNotExist
	}
}

func (br *Broccoli) Walk(root string, walkFn filepath.WalkFunc) error {
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
