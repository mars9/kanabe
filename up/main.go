package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const bufSize = 8 * 1024

var (
	modTime = flag.Bool("t", false, "use modification time to compare files")
	silent  = flag.Bool("s", false, "reporting through the exit status")
	report  = flag.Bool("c", false, "report changes instead of commands")

	src, dst string
)

func cmp(src, dst string) (bool, error) {
	fd1, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer fd1.Close()
	fd2, err := os.Open(dst)
	if err != nil {
		return false, err
	}
	defer fd2.Close()

	buf1, buf2 := make([]byte, bufSize), make([]byte, bufSize)
	offM, offN := int64(0), int64(0)
	eof := false

	for {
		m, err := fd1.ReadAt(buf1, offM)
		if err != nil {
			if err == io.EOF {
				eof = true
			} else {
				return false, err
			}
		}
		offM += int64(m)

		n, err := fd2.ReadAt(buf2, offN)
		if err != nil {
			if err == io.EOF {
				eof = true
			} else {
				return false, err
			}
		}
		offN += int64(n)
		if r := bytes.Compare(buf1, buf2); r != 0 {
			return false, nil
		}
		if eof {
			break
		}
	}
	return true, err
}

func cmpMeta(src, dst os.FileInfo) bool {
	if dst == nil {
		return true
	}
	smod := src.ModTime()
	dmod := dst.ModTime()
	if dmod.After(smod) {
		return false
	}
	return true
}

func rm(src string) {
	if *silent {
		os.Exit(1)
	}
	if *report {
		fmt.Printf("d %s\n", src)
		return
	}
	fmt.Printf("rm -rf %s\n", src)
}

func cp(src, dst string, isDir, append bool) {
	if *silent {
		os.Exit(1)
	}
	if *report {
		if append {
			fmt.Printf("a %s\n", dst)
		} else {
			fmt.Printf("c %s\n", dst)
		}
		return
	}
	if isDir {
		fmt.Printf("mkdir -p %s && dircp %s %s\n", dst, src, dst)
	} else {
		switch runtime.GOOS {
		case "plan9":
			fmt.Printf("fcp -x %s %s\n", src, dst)
		default:
			fmt.Printf("cp -p %s %s\n", src, dst)
		}
	}
}

func preUp(path string) string  { return strings.Replace(path, src, dst, 1) }
func preDel(path string) string { return strings.Replace(path, dst, src, 1) }

func update(srcPath string, src os.FileInfo, err error) error {
	dstPath := preUp(srcPath)

	if src.IsDir() {
		dst, err := os.Stat(dstPath)
		if err != nil {
			cp(srcPath, dstPath, true, true)
			return filepath.SkipDir
		}
		if !dst.IsDir() {
			rm(dstPath)
			cp(srcPath, dstPath, true, true)
			return filepath.SkipDir
		}
	} else {
		if *modTime {
			dst, err := os.Stat(dstPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
			if cmpMeta(src, dst) {
				cp(srcPath, dstPath, false, false)
			} else {
				cp(dstPath, srcPath, false, false)
			}
		} else {
			res, err := cmp(srcPath, dstPath)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}
			if !res {
				cp(srcPath, dstPath, false, false)
				return nil
			}
		}
	}
	return nil
}

func delete(dstPath string, dst os.FileInfo, err error) error {
	srcPath := preDel(dstPath)
	if _, err := os.Stat(srcPath); err != nil {
		rm(dstPath)
	}
	return nil
}

func init() {
	// (l)unix lacks dircp
	if _, err := exec.LookPath("dircp"); err != nil {
		gopath := strings.Split(os.Getenv("GOPATH"), ":")
		bin := filepath.Join(gopath[0], "bin", "dircp")
		err := os.MkdirAll(filepath.Join(gopath[0], "bin"), 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mkdirall %s/bin: %v", gopath[0], err)
			os.Exit(1)
		}
		f, err := os.Create(bin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", bin, err)
			os.Exit(1)
		}
		defer f.Close()
		if _, err = f.Write([]byte(dircp)); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", bin, err)
			os.Exit(1)
		}
		if err = os.Chmod(bin, 0755); err != nil {
			os.Remove(bin)
			fmt.Fprintf(os.Stderr, "chmod %s: %v", bin, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "# up installed dircp %q\n", bin)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] src dst\n", os.Args[0])
		fmt.Fprint(os.Stderr, usageMsg)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
	}
	src = flag.Arg(0)
	dst = flag.Arg(1)

	err := filepath.Walk(src, update)
	if err != nil {
		fmt.Fprintf(os.Stderr, "update %s %s: %v\n", src, dst, err)
		os.Exit(127)
	}
	err = filepath.Walk(dst, delete)
	if err != nil {
		fmt.Fprintf(os.Stderr, "delete %s %s: %v\n", dst, src, err)
		os.Exit(127)
	}
	os.Exit(0)
}

const usageMsg = `
Up prints shell commands needed to make dst like src. Both arguments
may name files or directories.

Flag -c makes up report changes instead of commands to update the
destination file. Each change is indicated by the initial of added,
deleted and changed followed by the file name, relative to the
destination directory.

Flag -s makes the tool silent, reporting through the exit status
non-zero only if src and dst differ.

If a file is inaccessible or missing, the exit status is '127'. The
exit status is not '0' only if src and dst differ.
`

const dircp = `#!/bin/sh
# dircp src dest - copy a tree with tar

case "$#" in
2)
	tar -C "$1" -pcf - . | tar -C "$2" -pxf -
	;;
*)
	echo Usage: dircp from to 1>&2
	;;
esac
`
