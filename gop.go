package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ryanuber/go-glob"
)

const (
	TempPrefix = "gop"
)

// ArgType are the command line arguments.
type ArgType struct {
	JarFile string   `arg:"positional,required,help:jar file"`
	Extra   []string `arg:"positional,help:extra parameters to pass to javap"`
	Glob    string   `arg:"--only,help:only process matching classes matching specified glob"`
	GlobNot string   `arg:"--not,help:exclusion glob"`
	Dex2Jar string   `arg:"--dex2jar,help:path to dex2jar executable"`
}

// processDex processes a dex file.
func processDex(args *ArgType, dexPath, tmpDir, origName string) error {
	var err error

	// get temp file name
	f, err := ioutil.TempFile(tmpDir, TempPrefix)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	err = os.Remove(f.Name())
	if err != nil {
		return err
	}

	// pass dex through dex2jar
	c := exec.Command(args.Dex2Jar, "-o", f.Name(), dexPath)
	_, err = c.Output()
	if err != nil {
		return err
	}

	// process jar
	return processJar(args, f.Name(), tmpDir, origName)
}

// processJar processes a jar file.
func processJar(args *ArgType, jarPath, tmpDir, origName string) error {
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
		case "jar", "apk", "aar", "dex":
			// open file in zip for reading
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()

			// open destination file
			tf, err := ioutil.TempFile(tmpDir, TempPrefix)
			if err != nil {
				return err
			}
			defer tf.Close()

			// write file
			_, err = io.Copy(tf, fr)
			if err != nil {
				return err
			}

			// if it's a dex, convert it
			if ext[1:] == "dex" {
				if args.Dex2Jar != "" {
					defer processDex(args, tf.Name(), tmpDir, path.Join(origName, f.Name))
				}
			} else {
				// process jar
				defer processJar(args, tf.Name(), tmpDir, path.Join(origName, f.Name))
			}

		case "class":
			// determine class name
			d := path.Dir(f.Name)
			n := path.Base(f.Name)
			n = n[:len(n)-len(ext)]
			if d != "" {
				n = strings.Replace(d, "/", ".", -1) + "." + n
			}

			// skip if classname doesn't match glob
			if args.Glob != "" && !glob.Glob(args.Glob, n) {
				continue
			}

			// skip if classname matches excluded glob
			if args.GlobNot != "" && glob.Glob(args.GlobNot, n) {
				continue
			}

			// execute javap on class
			v := append([]string{"-classpath", jarPath}, args.Extra...)
			v = append(v, n)
			c := exec.Command("javap", v...)
			out, err := c.Output()
			if err != nil {
				return err
			}

			// output data
			fmt.Fprintf(os.Stdout, "// class '%s' from '%s'\n", n, path.Join(origName, f.Name))
			if bytes.HasPrefix(out, []byte("Compiled from ")) {
				os.Stdout.Write([]byte("// "))
			}
			os.Stdout.Write(out)
			fmt.Fprint(os.Stdout, "\n")
		}
	}

	return nil
}

// findDex2Jar attempts to find dex2jar.
func findDex2Jar(args *ArgType) {
	dxHome := os.Getenv("DEXTOOLS_HOME")
	if dxHome == "" {
		return
	}

	// look for the dex2jar / dex2jar.sh file
	for _, n := range []string{"d2j-dex2jar.sh", "dex2jar.sh", "dex2jar"} {
		p := path.Join(dxHome, n)
		f, err := os.Stat(p)
		if err != nil {
			continue
		} else if f.IsDir() {
			continue
		}
		args.Dex2Jar = p
	}
}

func main() {
	var err error

	// parse args
	args := &ArgType{}
	arg.MustParse(args)

	// attempt to find dextools if path not specified
	if args.Dex2Jar == "" {
		findDex2Jar(args)
	}

	// replace ^ with - (issue with go-arg)
	for i, v := range args.Extra {
		args.Extra[i] = strings.Replace(v, "^", "-", -1)
	}

	// set default value for Extra
	if args.Extra == nil || len(args.Extra) == 0 {
		args.Extra = []string{"-p", "-constants"}
	}

	// create tempDir
	tmpDir, err := ioutil.TempDir(os.TempDir(), TempPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// process jar
	err = processJar(args, args.JarFile, tmpDir, args.JarFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
