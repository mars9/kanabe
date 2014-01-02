package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type perror string

func (e perror) Error() string { return string(e) }

func git(msg string) (err error) {
	cmd := exec.Command("git", "commit", "-m", msg)
	if err = cmd.Run(); err != nil {
		// not a git repository
		if err.Error() == "exit status 128" {
			return perror("not a git repo")
		}
		// nothing changed
		if err.Error() == "exit status 1" {
			fmt.Fprintf(os.Stdout, "%s: nothing changed\n", os.Args[0])
			return nil
		}
		return err
	}
	return nil
}

func mercurial(msg string) (err error) {
	cmd := exec.Command("hg", "commit", "-m", msg)
	if err = cmd.Run(); err != nil {
		// not a mercurial repository
		if err.Error() == "exit status 255" {
			return perror("not a hg repo")
		}
		// nothing changed
		if err.Error() == "exit status 1" {
			fmt.Fprintf(os.Stdout, "%s: nothing changed\n", os.Args[0])
			return nil
		}
		return err
	}
	return nil
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		fmt.Fprintf(os.Stderr, "Usage: %s [MESSAGE]\n", os.Args[0])
		os.Exit(2)
	}
	msg := fmt.Sprintf("%d", time.Now().Unix())
	if len(os.Args) == 2 {
		msg = os.Args[1]
	}
	err := mercurial(msg)
	if err != nil {
		if _, ok := err.(perror); ok {
			if err = git(msg); err != nil {
				if _, ok := err.(perror); ok {
					fmt.Fprintf(os.Stderr, "%s: not a hg/git repository\n", os.Args[0])
					os.Exit(0)
				}
				goto Error
			}
			os.Exit(0)
		}
		goto Error
	}
	os.Exit(0)

Error:
	fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
	os.Exit(1)

}
