package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "depx:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		typeName = flag.String("type", "", "implementation struct type to generate a Deps constructor for (required)")
		retType  = flag.String("return", "", "constructor return type (default \"*<type>\")")
		file     = flag.String("file", os.Getenv("GOFILE"), "source file containing the struct; defaults to $GOFILE under go:generate")
		outDir   = flag.String("dir", "", "output directory; defaults to the source file's directory")
	)
	flag.Parse()

	if *typeName == "" {
		return fmt.Errorf("-type is required (e.g. -type factory)")
	}
	if *file == "" {
		return fmt.Errorf("-file is required when not run under go:generate")
	}

	src, err := os.ReadFile(*file)
	if err != nil {
		return err
	}

	dir := *outDir
	if dir == "" {
		dir = filepath.Dir(*file)
	}

	spec, err := parse(*file, src, *typeName, *retType)
	if err != nil {
		return err
	}
	out, err := generate(spec)
	if err != nil {
		return err
	}
	outPath := filepath.Join(dir, strings.ToLower(*typeName)+"_deps.go")
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "depx: wrote", outPath)
	return nil
}
