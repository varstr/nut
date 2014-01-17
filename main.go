package main

import (
    "flag"
    "fmt"
)

const (
    doc = `The universe is in a nut.
    
Usage:

    nut <command> [<arguments>]

The commands are:
    make        make a nut for the target project
    open        deploy all packages according to an existed nut
    add         add a specific package to a nut
    `
)

func main() {
    flag.Parse()
    args := flag.Args()
    if len(args) == 0 {
        fmt.Println(doc)
        return
    }

    repo := args[1]

    switch args[0] {
    case "add":
        pkg := args[2]
        cmdAdd(repo, pkg)
    case "make":
        cmdMake(repo)
    case "open":
        cmdOpen(repo)
    default:
        fmt.Println(doc)
    }
}

func cmdMake(repo string) {
    nut := newNut(repo)
    nut.addDeps()
    nut.dump()
}

func cmdOpen(repo string) {
    nut := newNut(repo)
    nut.load()
    nut.deployDeps()
}

func cmdAdd(repo, pkg string) {
    nut := newNut(repo)
    nut.load()
    nut.addPkg(pkg)
    nut.dump()
}
