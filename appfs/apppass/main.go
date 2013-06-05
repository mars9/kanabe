package main

import (
	"crypto/rand"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	saltLen := flag.Int("len", 8, "salt length")
	hash := func(s string) string {
		h := sha1.New()
		h.Write([]byte(s))
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	salt := func() string {
		buf := make([]byte, *saltLen)
		n, err := io.ReadFull(rand.Reader, buf)
		if n != *saltLen || err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(1)
		}
		return fmt.Sprintf("%x", buf)
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s username password\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
	}

	s := salt()
	fmt.Printf("%s %s %s\n", flag.Arg(0), s, hash(s+flag.Arg(1)))
}
