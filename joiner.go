package main

import (
	"io"
	"os"
	"sort"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
)

func JoinFile(files []string, out string) error {
	//sort with file name or we will join files with wrong order
	sort.Strings(files)
	Printf("Start joining \n")
	bar := pb.StartNew(len(files)).Prefix(color.CyanString("Joining"))

	outf, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	defer outf.Close()
	if err != nil {
		return err
	}

	for _, f := range files {
		if err = copy(f, outf); err != nil {
			return err
		}
		bar.Increment()
	}
	bar.Finish()

	return nil
}

//this function split just to use defer
func copy(from string, to io.Writer) error {
	f, err := os.OpenFile(from, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	io.Copy(to, f)
	return nil
}
