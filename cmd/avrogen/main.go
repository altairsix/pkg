package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/altairsix/pkg/cmd/avrogen/app"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Options struct {
	Package   string
	UseStdout bool
}

var opts Options

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	v := cli.NewApp()
	v.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "pkg",
			Usage:       "Name of package; defaults to the name of the directory the avsc files are in",
			Destination: &opts.Package,
		},
		cli.BoolTFlag{
			Name:        "stdout",
			Usage:       "write all content to stdout instead of files",
			Destination: &opts.UseStdout,
		},
	}
	v.Action = Run
	v.Run(os.Args)
}

func Run(c *cli.Context) error {
	for _, arg := range c.Args() {
		err := WriteFile(arg)
		check(err)
	}

	return nil
}

func WriteFile(filename string) error {
	// read the contents of the avro file
	//
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to open file, %v", filename)
	}
	records, err := app.Parse(bytes.NewReader(data))
	if err != nil {
		return errors.Wrapf(err, "unable to parse avro file", filename)
	}

	w := os.Stdout
	if !opts.UseStdout {
		// create the file for writing
		//
		sourceFile := filename + ".go"
		f, err := os.Create(sourceFile)
		if err != nil {
			return errors.Wrapf(err, "unable to create file, %v", sourceFile)
		}
		defer f.Close()

		w = f
	}

	// determine package name
	//
	path, err := filepath.Abs(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to determine absolute path for file, %v", filename)
	}
	pkg := filepath.Base(filepath.Dir(path))
	fmt.Fprintf(w, "package %v\n", pkg)

	err = app.NewGenerator().Generate(w, records)
	if err != nil {
		return errors.Wrapf(err, "unable to generate content for file, %v", filename)
	}

	return nil
}
