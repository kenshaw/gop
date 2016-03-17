# About gop
gop is a simple utility written in [Go](https://golang.org/project) that makes
use of [javap](https://docs.oracle.com/javase/8/docs/technotes/tools/windows/javap.html)
to output all the class information contained within a jar/aar (ie, a zip
file). gop will also traverse the zip and extract the information for all
jars/classes contained within.

gop is mostly useful for quickly inspecting the contained java classes within a
jar.

# Usage

```sh
$ ./gop --help
usage: gop JARFILE

positional arguments:
  jarfile                jar file

options:
  --help, -h             display this help and exit
```
