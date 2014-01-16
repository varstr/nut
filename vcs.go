package main

import (
    "path/filepath"
)

type vcs struct {
    name string

    dlArgs,
    getRevArgs,
    getRemoteArgs,
    getRootArgs,
    toRevArgs []string
}

func args(a ...string) []string {
    return a
}

var (
    vcses = []*vcs {
        &vcs {
            name: "git",
            dlArgs: args("clone"),
            getRevArgs: args("log", "--pretty=format:%H", "-n", "1", "HEAD"),
            getRemoteArgs: args("config", "--get", "remote.origin.url"),
            getRootArgs: args("rev-parse", "--show-toplevel"),
            toRevArgs: args("checkout"),
        },
        &vcs {
            name: "hg",
            dlArgs: args("clone"),
            getRevArgs: args("log", "--template={node}", "--rev=."),
            getRemoteArgs: args("paths", "default"),
            getRootArgs: args("root"),
            toRevArgs: args("update"),
        },
    }
)

func selectVCS(name string) *vcs {
    for _, v := range vcses {
        if v.name == name {
            return v
        }
    }
    return nil
}

func detectVCS(dir string) *vcs {
    for _, v := range vcses {
        if root := tryRun(dir, v.name, v.getRootArgs...); len(root) > 0 {
            return v
        }
    }
    return nil
}

func (v *vcs) getAbsRoot(dir string) string {
    return tryRun(dir, v.name, v.getRootArgs...)
}

// TODO: re-consider later
func (v *vcs) getRelRoot(dir string) string {
    absRoot := run(dir, v.name, v.getRootArgs...)
    gopath := goPaths()

    for _, gopath := range gopath {
        prefix := filepath.Join(gopath, "src")
        if relRoot, err := filepath.Rel(prefix, absRoot); err == nil {
            return relRoot
        }
    }

    return ""
}

// TODO: deal with non-empty dir
func (v *vcs) download(remote, dir string) string {
    return run(".", v.name, append(v.dlArgs, remote, dir)...)
}

func (v *vcs) getRev(dir string) string {
    return run(dir, v.name, v.getRevArgs...)
}

func (v *vcs) getRemote(dir string) string {
    return run(dir, v.name, v.getRemoteArgs...)
}

func (v *vcs) toRev(dir, rev string) string {
    return run(dir, v.name, append(v.toRevArgs, rev)...)
}
