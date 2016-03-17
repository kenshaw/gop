# About gop
gop is a simple utility written in [Go](https://golang.org/project) that makes
use of [javap](https://docs.oracle.com/javase/8/docs/technotes/tools/windows/javap.html)
to output all the class information contained within a jar/aar (ie, a zip
file). gop will also traverse the zip and extract the information for all
jars/classes contained within.

gop is mostly useful for quickly inspecting the contained java classes within a
jar.

**NOTE:** `javap` must exist somewhere on your `$PATH` to use gop.

# Installation

Install in the usual Go way:
```sh
go get -u github.com/knq/gop
```

# Usage

```sh
$ gop --help
usage: gop JARFILE

positional arguments:
  jarfile                jar file

options:
  --help, -h             display this help and exit

$ gop /path/to/file.jar

# syntax highlight output
$ gop /path/to/file.jar |pygmentize -l java
```
