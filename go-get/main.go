package main

import (
	"os"
	"log"
	"os/exec"
	"strings"
	"github.com/pkg/errors"
	"fmt"
	"bufio"
	"path/filepath"
)

func main() {
	handleGoSubRepos()
}

var (
	gopath string
	out *wrapStdOut
	logger *log.Logger
)

func init() {
	out = &wrapStdOut{bf: bufio.NewWriter(os.Stdout)}
	logger = log.New(out, ">>> ", log.Ltime)

	paths := filepath.SplitList(os.Getenv("GOPATH"))
	if len(paths) == 0 {
		logger.Fatal("env:GOPATH not set")
	}
	gopath = filepath.ToSlash(paths[0])
	logger.Print("GOPATH: ", gopath)
}

type wrapStdOut struct {
	bf *bufio.Writer
}

func (out *wrapStdOut) Write(p []byte) (n int, err error) {
	out.bf.Flush()
	return out.bf.Write(p)
}

//golang.org/x/blog — the content and server program for blog.golang.org.
//golang.org/x/crypto — additional cryptography packages.
//golang.org/x/exp — experimental code (handle with care).
//golang.org/x/image — additional imaging packages.
//golang.org/x/mobile — libraries and build tools for Go on Android.
//golang.org/x/net — additional networking packages.
//golang.org/x/sys — for low-level interactions with the operating system.
//golang.org/x/talks — the content and server program for talks.golang.org.
//golang.org/x/text — packages for working with text.
//golang.org/x/tools — godoc, vet, cover, and other tools.
var goSubRepos = []string {
	"blog",
	"crypto",
	"exp",
	"image",
	"mobile",
	"net",
	"sys",
	"talks",
	"text",
	"tools",
	"lint",
}

func handleGoSubRepos() {
	for _, repo := range goSubRepos {
		pkg := MirrorPackage{
			ImportPath: fmt.Sprintf("golang.org/x/%s", repo),
			GitRemoteRepo: fmt.Sprintf("https://github.com/golang/%s", repo),
		}
		logger.Print("Merge pkg: ", pkg.ImportPath)
		if err := pkg.Merge(); err != nil {
			err = errors.WithMessage(err, fmt.Sprintf("pkg: %s do merge", pkg.ImportPath))
			logger.Fatal(err)
		}
	}
}

type MirrorPackage struct {
	ImportPath string
	GitRemoteRepo string
}

func (pkg MirrorPackage) ExistOnDisk() (dir string, exist bool, isGitRepo bool) {
	dir = fmt.Sprintf("%s/src/%s", gopath, pkg.ImportPath)
	if _, err := os.Stat(dir); err != nil {
		return
	}
	exist = true
	dirDotGit := strings.Join([]string{dir, ".git"}, "/")
	if _, err := os.Stat(dirDotGit); err != nil {
		return
	}
	isGitRepo = true
	return
}

// dir exist && contains .git, do 'git pull'
// dir exist && !contains .git,  remove dir; do 'git clone'
// !dir exist, do 'git clone'
func (pkg MirrorPackage) Merge() error {
	dir, exist, isGitRepo := pkg.ExistOnDisk()
	if exist {
		if isGitRepo {
			cmd := exec.Command("git", "pull", "--progress")
			cmd.Dir = dir
			cmd.Stdout = out
			cmd.Stderr = out
			if err := cmd.Run(); err != nil {
				return errors.WithMessage(err, "do 'git pull'")
			}
			return nil
		}
		if err := os.Remove(dir); err != nil {
			return errors.WithMessage(err, "remove existed dir")
		}
	}
	cmd := exec.Command("git", "clone", "--progress", pkg.GitRemoteRepo, dir)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return errors.WithMessage(err, "do 'git clone'")
	}
	return nil
}
