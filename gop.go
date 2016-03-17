package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alexflint/go-arg"
)

// ArgType are the command line arguments.
type ArgType struct {
	JarFile string `arg:"positional,required,help:jar file"`
}

// processJar processes a jar file.
func processJar(jarPath, tmpDir, origName string) error {
	var err error

	// open jar
	zr, err := zip.OpenReader(jarPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	// recursive scan for contained jars
	for _, f := range zr.File {
		// skip if directory
		if f.FileInfo().IsDir() {
			continue
		}

		// grab file extension
		ext := path.Ext(f.Name)
		if ext == "" {
			continue
		}

		// process file
		switch ext[1:] {
		case "jar", "aar":
			// open file in zip for reading
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()

			// open destination file
			tf, err := ioutil.TempFile(tmpDir, "gop")
			if err != nil {
				return err
			}
			defer tf.Close()

			// write file
			_, err = io.Copy(tf, fr)
			if err != nil {
				return err
			}

			// defer processing of jar
			defer processJar(tf.Name(), tmpDir, path.Join(origName, f.Name))

		case "class":
			// determine class name
			d := path.Dir(f.Name)
			n := path.Base(f.Name)
			n = n[:len(n)-len(ext)]
			if d != "" {
				n = strings.Replace(d, "/", ".", -1) + "." + n
			}

			// execute javap on class
			c := exec.Command("javap", "-classpath", jarPath, "-p", n)
			out, err := c.Output()
			if err != nil {
				return err
			}

			// output data
			fmt.Fprintf(os.Stdout, "// %s from %s\n// ", n, origName)
			os.Stdout.Write(out)
			fmt.Fprint(os.Stdout, "\n")
		}
	}

	return nil
}

func main() {
	var err error

	// parse args
	args := &ArgType{}
	arg.MustParse(args)

	// create tempDir
	tmpDir, err := ioutil.TempDir(os.TempDir(), "gop")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// process jar
	err = processJar(args.JarFile, tmpDir, args.JarFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
