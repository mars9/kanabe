// Derived from http://plan9.bell-labs.com/sources/plan9/sys/src/cmd/cmp.c
//
// Distributed under the Lucent Public License version 1.02
// http://plan9.bell-labs.com/plan9/license.html
//
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

const BUF = 65536

var (
	Lflag = flag.Bool("L", false, "print the line number of the differing byte")
	lflag = flag.Bool("l", false, "print the byte number and the differing\n"+
		"            bytes for each difference")
	silent = flag.Bool("s", false, "reporting through the exit status")
)

func seekoff(f *os.File, offset string) error {
	if offset != "" {
		n, err := strconv.ParseInt(offset, 0, 64)
		if err != nil {
			return err
		}
		if _, err = f.Seek(n, 0); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] file1 file2 [offset1 [offset2]]\n", os.Args[0])
		fmt.Fprint(os.Stderr, usageMsg)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(3)
	}
	flag.Parse()
	if flag.NArg() < 2 || flag.NArg() > 4 {
		flag.Usage()
	}

	fatal := func(err error) {
		fmt.Println("cmp: " + err.Error())
		os.Exit(127)
	}

	name1, name2 := flag.Arg(0), flag.Arg(1)
	f1, err := os.Open(name1)
	if err != nil {
		fatal(err)
	}
	defer f1.Close()

	f2, err := os.Open(name2)
	if err != nil {
		fatal(err)
	}
	defer f2.Close()

	if err = seekoff(f1, flag.Arg(2)); err != nil {
		fatal(err)
	}
	if err = seekoff(f2, flag.Arg(3)); err != nil {
		fatal(err)
	}

	var n int
	var nc uint64 = 1
	var l uint64 = 1
	len1, len2 := 0, 0
	buf1 := make([]byte, BUF)
	buf2 := make([]byte, BUF)
	for {
		n1, err := f1.Read(buf1)
		if err != nil {
			if err == io.EOF && n1 == 0 {
				break
			}
			fatal(err)
		}
		n2, err := f2.Read(buf2)
		if err != nil {
			if err == io.EOF && n2 == 0 {
				break
			}
			fatal(err)
		}

		n = n1
		if n1 > n2 {
			n = n2
		}
		len1 += n1
		len2 += n2

		for i, ac := range buf1[0:n] {
			if ac == '\n' {
				l++
			}
			bc := buf2[i]
			if ac != bc {
				if *silent {
					os.Exit(1)
				}
				if !*lflag {
					if *Lflag {
						fmt.Printf("%s %s differ: char %d line %d\n",
							name1, name2, nc+uint64(i), l)
						os.Exit(1)
					} else {
						fmt.Printf("%s %s differ: char %d\n",
							name1, name2, nc+uint64(i))
						os.Exit(1)
					}
				}
				fmt.Printf("%d 0x%x 0x%x\n", nc+uint64(i), ac, bc)
			}
		}

		nc += uint64(n)
	}

	if len1 == len2 {
		os.Exit(0)
	}
	if !*silent {
		if len1 > len2 {
			fmt.Printf("EOF on %s after %d bytes\n", name2, nc-1)
		} else {
			fmt.Printf("EOF on %s after %d bytes\n", name1, nc-1)
		}
	}
	os.Exit(2)
}

const usageMsg = `
Cmp compares the two files and prints a message if the contents
differ.

If offsets are given, comparison starts at the designated byte
position of the corresponding file. Offsets that begin with 0x are
hexadecimal; with 0, octal; with anything else, decimal.

If a file is inaccessible or missing, the exit status is '127'. If the
files are the same, the exit status is '0'. If they are the same
except that one is longer than the other, the exit status is '2'.
Otherwise cmp reports the position of the first disagreeing byte and
the exit status is '1'.
`
