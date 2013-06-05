package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/russross/blackfriday"
)

func main() {
	latex := flag.Bool("latex", false, "generate LaTeX output instead of HTML")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s < input > output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
		os.Exit(1)
	}

	var renderer blackfriday.Renderer
	if *latex {
		renderer = blackfriday.LatexRenderer(0)
	} else {
		htmlFlags := 0
		htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
		htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
		htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
		renderer = blackfriday.HtmlRenderer(0, "", "")
	}

	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	os.Stdout.Write(blackfriday.Markdown(data, renderer, extensions))
}
