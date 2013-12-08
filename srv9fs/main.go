// +build plan9

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"bitbucket.org/fhs/go-plan9-auth/auth"
	"code.google.com/p/go.crypto/ssh"
)

var (
	user   string
	host   string
	port   string
	path   = flag.String("u", "$HOME/bin/u9fs", "u9fs binary on the remote system")
	chatty = flag.Bool("D", false, "write chatty output to the log file")
)

type passwd string

func (p passwd) Password(usr string) (string, error) {
	param := "proto=pass service=ssh role=client server=%s user=%s"
	_, passwd, err := auth.GetUserPassword(auth.GetKey, param, host, usr)
	return passwd, err
}

func parseArg(system string) (user string, host string, port string, err error) {
	port = "22"
	if strings.Contains(system, "@") {
		tmp := strings.Split(system, "@")
		if len(tmp) != 2 || tmp[1] == "" {
			err = fmt.Errorf("no ssh2 host")
			return
		}
		user = tmp[0]
		tmp = strings.Split(tmp[1], ":")
		if len(tmp) == 2 {
			port = tmp[1]
		}
		host = tmp[0]
	} else {
		err = fmt.Errorf("no ssh2 user")
		return
	}
	return
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] user@system[:port]\n", os.Args[0])
	fmt.Fprint(os.Stderr, usageMsg)
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var err error
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) != 1 {
		usage()
	}

	arg := flag.Arg(0)
	user, host, port, err = parseArg(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v", arg, err)
		os.Exit(2)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(passwd(user)),
		},
	}

	addr := host + ":" + port
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial %s: %v", addr, err)
		os.Exit(1)
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "session: %v", err)
		os.Exit(1)
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	args := " -nz"
	if *chatty {
		args = " -Dnz"
	}
	cmd := *path + args + " -a none -u $USER -l $HOME/u9fs.log"
	if err := session.Run(cmd); err != nil {
		fmt.Fprintf(os.Stderr, "run %s: %v", cmd, err)
		os.Exit(1)
	}
}

const usageMsg = `
Srv9fs connects to a remote Unix system via SSH and starts u9fs(4).
The -u option specifies the path to the u9fs binary on the remote
system.

Example:
  srv -s 5 -e 'srv9fs user@example.com' example /n/example
`
