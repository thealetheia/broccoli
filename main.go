package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	flagInput     = flag.String("src", "public", "")
	flagOutput    = flag.String("o", "", "")
	flagVariable  = flag.String("var", "br", "")
	flagInclude   = flag.String("include", "", "")
	flagExclude   = flag.String("exclude", "", "")
	flagOptional  = flag.Bool("opt", false, "")
	flagGitignore = flag.Bool("gitignore", false, "")
	flagQuality   = flag.Int("quality", 11, "")

	verbose = flag.Bool("v", false, "")
)

const (
	constInput = "public"
)

const help = `Usage: broccoli [options]

Broccoli uses brotli compression to embed a virtual file system in Go executables.

Options:
	-src folder[,file,file2]
		The input files and directories, "public" by default.
	-o
		Name of the generated file, follows input by default.
	-var=br
		Name of the exposed variable, "br" by default.
	-include *.html,*.css
		Wildcard for the files to include, no default.
	-exclude *.wasm
		Wildcard for the files to exclude, no default.
	-opt
		Optional decompression: if enabled, files will only be decompressed
		on the first time they are read.
	-gitignore
		Enables .gitignore rules parsing in each directory, disabled by default.
	-quality [level]
		Brotli compression level (0-11), the highest by default.

Generate a broccoli.gen.go file with the variable broccoli:
	//go:generate broccoli -src assets -o broccoli -var broccoli

Generate a regular public.gen.go file, but include all *.wasm files:
	//go:generate broccoli -src public -include="*.wasm"`

var goIdentifier = regexp.MustCompile(`^\p{L}[\p{L}0-9_]*$`)

func main() {
	log.SetFlags(0)
	log.SetPrefix("broccoli: ")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, help)
	}

	flag.Parse()
	if len(os.Args) == 1 {
		flag.Usage()
	}

	var inputs []string
	if flagInput == nil {
		inputs = []string{constInput}
	} else {
		inputs = strings.Split(*flagInput, ",")
	}

	output := *flagOutput
	if output == "" {
		output = inputs[0]
	}
	if !strings.HasSuffix(output, ".gen.go") {
		output = strings.Split(output, ".")[0] + ".gen.go"
	}

	variable := *flagVariable
	if !goIdentifier.MatchString(variable) {
		log.Fatalln(variable, "is not a valid Go identifier")
	}

	includeGlob := *flagInclude
	excludeGlob := *flagExclude
	if includeGlob != "" && excludeGlob != "" {
		log.Fatal("mutually exclusive options -include and -exclude found")
	}

	quality := *flagQuality
	if quality < 1 || quality > 11 {
		log.Fatalf("unsupported compression level %d (1-11)\n", quality)
	}

	g := Generator{
		inputFiles:   inputs,
		includeGlob:  includeGlob,
		excludeGlob:  excludeGlob,
		useGitignore: *flagGitignore,
		quality:      quality,
	}

	g.parsePackage()

	bundle, err := g.generate()
	if err != nil {
		log.Fatal(err)
	}

	code := fmt.Sprintf(template,
		time.Now().Format(time.RFC3339),
		g.pkg.name, variable, *flagOptional, bundle)

	err = ioutil.WriteFile(output, []byte(code), 0644)
	if err != nil {
		log.Fatalf("could not write to %s: %v\n", output, err)
	}
}
