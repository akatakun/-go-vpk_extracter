package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/akatakun/vpk-1"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: vpkverify [file1.vpk] [file2.vpk] [file3.vpk]\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	verbose := flag.Bool("v", false, "print the names of files even if they are valid")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
	}

	hadError := false

	for _, name := range flag.Args() {
		var opener vpk.Opener
		if strings.HasSuffix(name, "_dir.vpk") {
			opener = vpk.MultiVPK(name[:len(name)-len("_dir.vpk")])
		} else {
			opener = vpk.SingleVPK(name)
		}
		v, err := vpk.Open(opener)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			hadError = true
			continue
		}
		base := name[:len(name)-len(filepath.Ext(name))]

		for _, rel := range v.Paths() {
			dstDirpath := filepath.Join(base, filepath.Dir(rel))
			err = os.MkdirAll(dstDirpath, os.ModePerm)
			if err != nil {
				panic(err)
			}
			dstPath := filepath.Join(base, rel)
			dst, err := os.Create(dstPath)
			if err != nil {
				panic(err)
			}
			defer dst.Close()

			r, err := v.Entry(rel).Open()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: open %s: %v\n", name, rel, err)
				hadError = true
				continue
			}
			_, err = io.Copy(dst, r)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: read %s: %v\n", name, rel, err)
				hadError = true
				r.Close()
				continue
			}
			err = r.Close()
			if err != nil {
				fmt.Printf("%s: %s: %v\n", name, rel, err)
				hadError = true
			} else if *verbose {
				fmt.Printf("%s: %s is valid\n", name, rel)
			}
		}
	}

	if hadError {
		os.Exit(1)
	}
}
