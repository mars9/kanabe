package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	gopath   = os.Getenv("GOPATH")
	origRepo = "github.com/boltdb/bolt"
)

func fetch() (string, []byte, error) {
	tmp, err := ioutil.TempDir("", "bolddb")
	if err != nil {
		return "", nil, err
	}

	url := "https://" + origRepo
	log.Printf("clone %s", url)
	cmd := exec.Command("git", "clone", url, tmp)
	if err = cmd.Run(); err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	cmd = exec.Command("git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = tmp
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err = cmd.Run(); err != nil {
		return "", nil, err
	}
	return tmp, buf.Bytes(), nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s destination-path\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	if os.Args[1] == "-h" {
		usage()
	}

	tmp, rev, err := fetch()
	defer os.RemoveAll(tmp)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.MkdirAll(os.Args[1], 0755); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(path.Join(os.Args[1], "REV"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write(rev); err != nil {
		log.Fatal(err)
	}

	dir, err := os.Open(tmp)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		src := path.Join(tmp, f.Name())
		if strings.HasSuffix(src, "_test.go") {
			continue
		}

		sf, err := os.Open(src)
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasSuffix(src, ".go") || f.Name() == "LICENSE" {
			dst := path.Join(os.Args[1], f.Name())
			df, err := os.Create(dst)
			if err != nil {
				sf.Close()
				log.Fatal(err)
			}

			log.Printf("update %s", dst)
			if _, err = io.Copy(df, sf); err != nil {
				sf.Close()
				df.Close()
				log.Fatal(err)
			}
			sf.Close()
			df.Close()
			continue
		}
	}
}
