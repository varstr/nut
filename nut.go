package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	nutFile     = ".nut"
	nutFilePerm = 0666
	dirPerm     = 0755
)

type Repo struct {
	Root string
	VCS  string   `json:"vcs"`
	Rev  string   `json:"rev"`
	Pkgs []string `json:"pkgs"`
}

type Nut struct {
	Path string           `json:"path"`
	Deps map[string]*Repo `json:"deps"`
}

func newNut(path string) *Nut {
	return &Nut{
		Path: path,
		Deps: make(map[string]*Repo),
	}
}

func (n *Nut) addPkg(pkg string) {
	dir := goListDir(pkg)
	if len(dir) == 0 {
		return
	}
	vcs := detectVCS(dir)
	if vcs == nil {
		return
	}

	remote := vcs.getRemote(dir)
	repo, ok := n.Deps[remote]
	if ok {
		repo.Pkgs = append(repo.Pkgs, pkg)
		return
	}

	n.Deps[remote] = &Repo{
		Root: vcs.getRelRoot(dir),
		VCS:  vcs.name,
		Rev:  vcs.getRev(dir),
		Pkgs: []string{pkg},
	}
}

func (n *Nut) addDeps() {
	deps := goListDeps(n.Path)
	for _, dep := range deps {
		if strings.HasPrefix(dep, n.Path) {
			continue
		}
		n.addPkg(dep)
	}
}

func (n *Nut) deployDeps() {
	gopath := goPaths()[0]

	for remote, repo := range n.Deps {
		vcs := selectVCS(repo.VCS)
		absPath := filepath.Join(gopath, "src", repo.Root)
		if _, err := os.Stat(absPath); err == nil {
			if r := vcs.getRemote(absPath); r != remote {
				if err = os.RemoveAll(absPath); err != nil {
					exitOnError("rm "+absPath, err)
				}
				if err = os.MkdirAll(absPath, dirPerm); err != nil {
					exitOnError("mkdir "+absPath, err)
				}
				vcs.download(remote, absPath)
			}
		} else if os.IsNotExist(err) {
			vcs.download(remote, absPath)
		} else {
			exitOnError("stat "+absPath, err)
		}
		vcs.toRev(absPath, repo.Rev)
	}

	for _, repo := range n.Deps {
		for _, pkg := range repo.Pkgs {
			goInstall(pkg)
		}
	}
}

func (n *Nut) dump() {
	dir := goListDir(n.Path)

	buffer, err := json.MarshalIndent(n, "", " ")
	if err != nil {
		exitOnError("json marshal", err)
	}
	if err = os.Chdir(dir); err != nil {
		exitOnError("chdir "+dir, err)
	}
	if err = ioutil.WriteFile(nutFile, buffer, nutFilePerm); err != nil {
		exitOnError("write file "+nutFile, err)
	}
}

func (n *Nut) load() {
	dir := goListDir(n.Path)

	if err := os.Chdir(dir); err != nil {
		exitOnError("chdir "+dir, err)
	}
	buffer, err := ioutil.ReadFile(nutFile)
	if err != nil {
		exitOnError("read file "+nutFile, err)
	}
	if err = json.Unmarshal(buffer, n); err != nil {
		exitOnError("json unmarshal", err)
	}
}
