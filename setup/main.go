package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/go-github/github"
)

var (
	externalPackages = []string{
		"github.com/golang/protobuf/protoc-gen-go",

		"golang.org/x/crypto/...",
		"golang.org/x/net/...",
		"golang.org/x/tools/cmd/goimports",
		"golang.org/x/tools/cmd/stringer",
		"golang.org/x/tools/cmd/cover",
		"golang.org/x/tools/cmd/godoc",

		"github.com/golang/lint/golint",

		"github.com/boltdb/bolt",
	}

	kanabeCommands = []string{
		"crypt",
		"kanabe/codesearch",
		"kanabe/jfmt",
	}
)

func listRepos(user string) ([]string, error) {
	client := github.NewClient(nil)

	opt := &github.RepositoryListOptions{
		Type:      "owner",
		Sort:      "updated",
		Direction: "desc",
	}
	repos, _, err := client.Repositories.List(user, opt)
	if err != nil {
		return nil, err
	}

	var r []string
	for _, repo := range repos {
		if repo.Name != nil && repo.Language != nil {
			if strings.ToLower(*repo.Language) == "go" {
				r = append(r, *repo.Name)
			}
		}
	}
	return r, nil
}

func BuildExternalPackages() (err error) {
	for _, pkg := range externalPackages {
		fmt.Printf("get %s\n", pkg)
		cmd := exec.Command("go", "get", "-u", pkg)

		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	return err
}

func CloneRepositories(user string, repos []string) (err error) {
	dir := path.Join(GOPATH, "src", "github.com", user)
	if err = os.RemoveAll(dir); err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	for _, repo := range repos {
		addr := fmt.Sprintf("git@github.com:%s/%s.git", user, repo)
		fmt.Printf("clone %s\n", addr)

		cmd := exec.Command("git", "clone", addr)
		cmd.Dir = dir

		//cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	return err
}

func BuildCommands(user string) (err error) {
	dir := path.Join(GOPATH, "src", "github.com", user)
	for _, command := range kanabeCommands {
		fmt.Printf("build %s\n", command)
		cmd := exec.Command("go", "install", "./...")
		cmd.Dir = path.Join(dir, command)

		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("build private repo: %v", err)
		}
	}
	return err
}

var GOPATH = os.Getenv("GOPATH")

func main() {
	if GOPATH == "" {
		log.Fatalf("GOPATH environment variable not set")
	}

	var user = flag.String("user", "mars9", "github username")
	flag.Parse()

	binDir := path.Join(GOPATH, "bin")
	pkgDir := path.Join(GOPATH, "pkg")
	fmt.Printf("purging %s\n", binDir)
	if err := os.RemoveAll(binDir); err != nil {
		log.Fatalf("purging %s: %v", binDir, err)
	}
	if err := os.MkdirAll(binDir, 0750); err != nil {
		log.Fatalf("purging %s: %v", binDir, err)
	}
	fmt.Printf("purging %s\n", pkgDir)
	if err := os.RemoveAll(pkgDir); err != nil {
		log.Fatalf("purging %s: %v", pkgDir, err)
	}

	if err := BuildExternalPackages(); err != nil {
		log.Fatalf("build external packages: %v", err)
	}

	repos, err := listRepos(*user)
	if err != nil {
		log.Fatalf("list repositories: %v", err)
	}
	if err := CloneRepositories(*user, repos); err != nil {
		log.Fatalf("clone repositories: %v", err)
	}

	if err := BuildCommands(*user); err != nil {
		log.Fatalf("build commands: %v", err)
	}
}
