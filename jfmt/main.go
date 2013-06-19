package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s < input > output\n", os.Args[0])
		os.Exit(2)
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
		os.Exit(1)
	}
	buf := bytes.Buffer{}
	json.Indent(&buf, data, "", "  ")
	if data[len(data)-1] != '\n' {
		buf.WriteByte('\n')
	}
	os.Stdout.Write(buf.Bytes())
}
