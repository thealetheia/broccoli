package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

var (
	flagInput    = flag.String("i", path.Join(".", "public"), "")
	flagOutput   = flag.String("o", "", "")
	flagVariable = flag.String("var", "", "")
	flagExclude  = flag.String("exclude", "", "")
	flagQuality  = flag.Int("quality", 11, "")
)

const help = `Usage: broccoli [options]

Broccoli uses brotli compression to embed a virtual file system in Go executables.

Options:
  -i
	The input directory, "public" by default.
  -o
	Name of the generated file, input folder name by default.
  -var
	Name of the broccoli variable, either input folder name, or
	output file's base filename by default.
  -exclude
	Wildcard for the files to exclude, no default.
  -quality [level]
	Brotli compression level (0-11), highest by default.

Generate a broccoli.gen.go file with the variable broccoli:
	//go:generate broccoli -i assets -o broccoli -var broccoli

Generate a regular public.gen.go file, but exclude all *.exe files:
	//go:generate broccoli -i public -exclude="*.exe"
`

func main() {
	log.SetFlags(0)
	log.SetPrefix("broccoli: ")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, help)
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
	}

}
