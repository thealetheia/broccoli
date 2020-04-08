package broccoli
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var virtualDirs []string
var realDirs []string
func testBroccoli(br *Broccoli) {
	br.Walk("/testdata", func(path string, info os.FileInfo, err error) error {
		virtualDirs = append(virtualDirs, path)
		fmt.Println(path)

		return nil
	})
	fmt.Println(virtualDirs)
	fmt.Println(realDirs)
}
func TestBroccoli(t *testing.T) {
	var files []*File
	filepath.Walk("testdata", func(fs string, info os.FileInfo, err error) error {
		fl, err := NewFile(fs)
		realDirs = append(realDirs, fs)
		if err != nil && !strings.Contains(err.Error(), "is a directory") {
			log.Fatal(err)
			return err
		}
		if fl != nil {
			files = append(files, fl)
		}
		return nil
	})
	bytes, err := Pack(files, 6)

	if err != nil {
		log.Fatal(err)
		return
	}
	br, err := New(bytes)
	if err != nil {
		log.Fatal(err)
		return
	}
	testBroccoli(br)
}