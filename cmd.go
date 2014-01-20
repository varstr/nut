package main

import (
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

const (
    depTpl = `{{join .Deps " "}}`
    dirTpl = `{{if not .Standard}}{{.Dir}}{{end}}`

    errFmt = "run `%s` fail\n"
)

func exitOnError(caller string, extra ...interface{}) {
    extraFmt := strings.Repeat("%v\n", len(extra))
    os.Stderr.WriteString(fmt.Sprintf(errFmt, caller) + fmt.Sprintf(extraFmt, extra...))
    os.Exit(1)
}

func runCmd(ignoreErr bool, dir, name string, args ...string) string {
    err := os.Chdir(dir)
    if err != nil && !ignoreErr {
        exitOnError("chdir "+dir, err)
    }

    stdout := new(bytes.Buffer)
    stderr := new(bytes.Buffer)
    cmd := exec.Command(name, args...)
    cmd.Stdout = stdout
    cmd.Stderr = stderr

    if err = cmd.Run(); err != nil && !ignoreErr {
        exitOnError(strings.Join(cmd.Args, " "), err, stderr.String())
    }
    out := strings.TrimSpace(stdout.String())
    return out
}

func run(dir, name string, args ...string) string {
    return runCmd(false, dir, name, args...)
}

func tryRun(dir, name string, args ...string) string {
    return runCmd(true, dir, name, args...)
}

func goEnv(v string) string {
    return run(".", "go", "env", v)
}

func goPaths() []string {
    return strings.Split(goEnv("GOPATH"), ":")
}

func goList(tpl, path string) string {
    return run(".", "go", "list", "-f", tpl, path)
}

func goListDeps(path string) []string {
    return strings.Split(goList(depTpl, path), " ")
}

func goListDir(path string) string {
    return goList(dirTpl, path)
}

func goInstall(path string) string {
    return run(".", "go", "install", path)
}
