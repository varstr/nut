package main

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

const (
    nutFile = ".nut.json"
)

type Repo struct {
    Root string
    VCS string `json:"vcs"`
    Rev string `json:"rev"`
    Pkgs []string `json:"pkgs"`
}

type Nut struct {
    Path string `json:"path"`
    Deps map[string]*Repo `json:"deps"`
}

func newNut(path string) *Nut {
    return &Nut {
        Path: path,
        Deps: make(map[string]*Repo),
    }
}

func (n *Nut) addPkg(pkg string) {
    dir := goGetDir(pkg)
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

    n.Deps[remote] = &Repo {
        Root: vcs.getRelRoot(dir),
        VCS: vcs.name,
        Rev: vcs.getRev(dir),
        Pkgs: []string{pkg},
    }
}

func (n *Nut) addDeps() {
    deps := goGetDeps(n.Path)
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
        vcs.download(remote, absPath)
        vcs.toRev(absPath, repo.Rev)
        for _, pkg := range repo.Pkgs {
            goInstall(pkg)
        }
    }
}

func (n *Nut) dump() {
    dir := goGetDir(n.Path)

    buffer, err := json.MarshalIndent(n, "", " ")
    if err != nil {
        exitOnError("json marshal", err)
    }
    if err = os.Chdir(dir); err != nil {
        exitOnError("chdir "+dir, err)
    }
    if err = ioutil.WriteFile(nutFile, buffer, os.ModePerm); err != nil {
        exitOnError("write file "+nutFile, err)
    }
}

func (n *Nut) load() {
    dir := goGetDir(n.Path)

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

