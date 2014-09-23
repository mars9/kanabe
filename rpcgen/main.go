package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

var (
	funcReg  = regexp.MustCompile(`[A-Za-z0-9_]*\([A-Za-z0-9_]*\)`)
	respReg  = regexp.MustCompile(`\([A-Za-z0-9_]*\)`)
	cleanReg = regexp.MustCompile(`[A-Za-z0-9_]*\(`)
	nameReg  = regexp.MustCompile(`\([A-Za-z0-9_]*\)`)
)

type Package struct {
	Name    string
	Service []Service
}

type Service struct {
	Name     string
	Request  string
	Response string
	Index    int
}

func Parse(r io.Reader) (pack Package, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		if scanner.Text() == "package" {
			if !scanner.Scan() {
				return pack, fmt.Errorf("package syntax: %s", scanner.Text())
			}
			pack.Name = scanner.Text()
			pack.Name = pack.Name[:len(pack.Name)-1]
		}
		if scanner.Text() == "rpc" {
			i := 0
			s := Service{}
			for scanner.Scan() {
				switch i {
				case 0:
					if !funcReg.MatchString(scanner.Text()) {
						log.Fatalf("rpc syntax error: %s", scanner.Text())
					}

					s.Name = nameReg.ReplaceAllString(scanner.Text(), "")
					s.Request = cleanReg.ReplaceAllString(scanner.Text(), "")
					s.Request = s.Request[:len(s.Request)-1]
				case 2:
					if !respReg.MatchString(scanner.Text()) {
						return pack, fmt.Errorf("rpc syntax error: %s", scanner.Text())
					}

					s.Response = cleanReg.ReplaceAllString(scanner.Text(), "")
					s.Response = s.Response[:len(s.Response)-1]
				}
				i++
				if i >= 3 {
					break
				}
			}
			s.Index = len(pack.Service) + 1
			pack.Service = append(pack.Service, s)
		}
	}
	if err := scanner.Err(); err != nil {
		return pack, err
	}
	return pack, err
}

func Formatter(buf *bytes.Buffer) ([]byte, error) {
	cmd := exec.Command("gofmt")
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	if _, err = in.Write(buf.Bytes()); err != nil {
		return nil, err
	}
	in.Close()

	formated, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, err
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}
	return formated, err
}

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s protobuf.proto\n", os.Args[0])
		os.Exit(2)
	}

	pack, err := Parse(os.Stdin)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("proto").Parse(templ)
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buf, pack); err != nil {
		panic(err)
	}

	data, err := Formatter(buf)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(data)
}

const templ = `package {{.Name}}

import (
	"errors"
	"sync"

	pb "code.google.com/p/goprotobuf/proto"
)

const (
	{{range .Service}}{{.Name}}Service = {{.Index}}
	{{end}}
)

var (
	ErrInvalidArgs  = errors.New("invalid arguments")
	ErrInvalidReply = errors.New("invalid reply")
)

{{range .Service}}
type {{.Name}}Handler struct{
	argsPool  sync.Pool
	replyPool sync.Pool
} 

func (s *{{.Name}}Handler) service(args *{{.Request}}, reply *{{.Response}}) error {
	return nil
}

func (s *{{.Name}}Handler) Service(args pb.Message, reply pb.Message) error {
	req, ok := args.(*{{.Request}})
	if !ok {
		return ErrInvalidArgs 
	}
	resp, ok := reply.(*{{.Response}})
	if !ok {
		return ErrInvalidReply
	}
	return s.service(req, resp)
}

func (s *{{.Name}}Handler) Get() (args pb.Message, reply pb.Message) {
	return s.argsPool.Get().(*{{.Request}}), s.replyPool.Get().(*{{.Response}})
}

func (s *{{.Name}}Handler) Put(args pb.Message, reply pb.Message) {
	args.Reset()
	reply.Reset()
	s.argsPool.Put(args)
	s.replyPool.Put(reply)
}

{{end}}
`
