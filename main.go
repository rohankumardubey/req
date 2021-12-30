package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/andrewpillar/req/syntax"
	"github.com/andrewpillar/req/token"
)

func files() ([]string, error) {
	dir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	ents, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	fnames := make([]string, 0, len(ents))

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		if fname := ent.Name(); strings.HasSuffix(fname, ".req") {
			fnames = append(fnames, filepath.Join(dir, fname))
		}
	}

	sort.Strings(fnames)
	return fnames, nil
}

func errh(errs chan error) func(token.Pos, string) {
	return func(pos token.Pos, msg string) {
		pos.File = filepath.Base(pos.File)

		errs <- errors.New(pos.String() + " - " + msg)
	}
}

func main() {
	argv0 := os.Args[0]

	fnames, err := files()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", argv0, err)
		os.Exit(1)
	}

	sems := make(chan struct{}, runtime.GOMAXPROCS(0)+10)
	errs := make(chan error)

	var wg sync.WaitGroup
	wg.Add(len(fnames))

	for _, fname := range fnames {
		go func(fname string) {
			sems <- struct{}{}
			defer func() {
				wg.Done()
				<-sems
			}()

			syntax.ParseFile(fname, errh(errs))
		}(fname)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	var errc int

	errmax := 50

	for err := range errs {
		if errc < errmax {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		errc++
	}

	if errc > 0 {
		if errc > errmax {
			fmt.Fprintf(os.Stderr, "%s: too many errors\n", argv0)
		}
		os.Exit(1)
	}
}