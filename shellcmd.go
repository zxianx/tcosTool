package main

import (
    "bytes"
    "errors"
    "fmt"
    "io/ioutil"
    "os/exec"
)

func WaitForStdOut(cmd *exec.Cmd) (string, error) {
    stdout, _ := cmd.StdoutPipe()
    if err := cmd.Start(); err != nil {
        fmt.Println("Execute failed when Start:" + err.Error())
        return err.Error(), err
    }

    out_bytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        return err.Error(), err
    }
    stdout.Close()
    if err := cmd.Wait(); err != nil {
        return err.Error(), err
    }
    return string(out_bytes), err
}

func ExecCommand(strCommand string) (stdOut string, stdErr string, pid int, err error) {
    var res []byte
    var stderr bytes.Buffer
    cmd := exec.Command("/bin/bash", "-c", strCommand)
    cmd.Stderr = &stderr
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Println("Execute failed when Start:" + err.Error())
        return "", "", 0, err
    }
    defer stdout.Close()
    err = cmd.Start()
    if err != nil {
        return "", stderr.String(), 0, err
    }
    pid = cmd.Process.Pid
    res, err = ioutil.ReadAll(stdout)
    if err != nil {
        return "", stderr.String(), pid, err
    }
    err = cmd.Wait()
    return string(res), stderr.String(), pid, err
}

func ExecCommandWithoutResult(strCommand string) (err error) {
    _, stdErr, _, err := ExecCommand(strCommand)
    if err != nil {
        return err
    }
    if stdErr != "" {
        return errors.New("std err:" + stdErr)
    }
    return
}
